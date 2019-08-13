package log

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/cachecashproject/go-cachecash/common"
	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	// Heartbeat -- if set to true, will attempt to deliver the logs every
	// TickInterval. Set this to false to not deliver logs at all, and instead just
	// write them. Useful for testing.
	Heartbeat = true

	// BackoffCap is the maximum time we will wait before attempting to deliver logs
	BackoffCap = 5 * time.Minute

	// BackoffGranularity is the granularity by which backoff is calcluated
	BackoffGranularity = time.Second

	// DefaultTickInterval is a singleton for how long to wait for a log file to fill
	// before delivering it.
	DefaultTickInterval = time.Second
)

// Client is a logging client that uses grpc to send a structured log.
type Client struct {
	service string
	logDir  string

	logFile *os.File
	logLock sync.Mutex

	ourLogger *logrus.Logger

	tickerMutex    sync.Mutex
	tickerDuration time.Duration

	logPipeClient LogPipeClient

	heartbeatCancel context.CancelFunc

	errorMutex sync.RWMutex
	Error      error
}

// NewClient creates a new client.
func NewClient(serverAddress, service, logDir string, debug bool) (*Client, error) {
	if err := os.MkdirAll(logDir, 0700); err != nil {
		return nil, err
	}

	c := &Client{
		service:   service,
		logDir:    logDir,
		ourLogger: logrus.New(),
	}

	if debug {
		c.ourLogger.SetLevel(logrus.DebugLevel)
	}

	if Heartbeat {
		conn, err := common.GRPCDial(serverAddress)
		if err != nil {
			return nil, err
		}

		c.logPipeClient = NewLogPipeClient(conn)

		ctx, cancel := context.WithCancel(context.Background())
		c.heartbeatCancel = cancel

		c.adjustTicker(DefaultTickInterval)
		go c.heartbeat(ctx)
	}

	return c, c.makeLog(true)
}

func (c *Client) adjustTicker(dur time.Duration) {
	c.tickerMutex.Lock()
	defer c.tickerMutex.Unlock()
	c.tickerDuration = dur
}

// Close closes any logfile and connections
func (c *Client) Close() error {
	c.tickerMutex.Lock()
	defer c.tickerMutex.Unlock()

	if c.heartbeatCancel != nil {
		c.heartbeatCancel()
	}

	c.logLock.Lock()
	defer c.logLock.Unlock()

	if c.logFile != nil {
		lf := c.logFile
		c.logFile = nil
		if err := lf.Sync(); err != nil {
			return err
		}

		if err := lf.Close(); err != nil {
			return err
		}
	}

	if err := c.makeLog(false); err != nil {
		return err
	}

	c.errorMutex.RLock()
	defer c.errorMutex.RUnlock()
	return c.Error
}

func (c *Client) heartbeat(ctx context.Context) {
	for {
		c.tickerMutex.Lock()
		after := time.After(c.tickerDuration)
		c.tickerMutex.Unlock()
		select {
		case <-ctx.Done():
			return
		case <-after:
			if err := c.makeLog(true); err != nil {
				c.errorMutex.Lock()
				defer c.errorMutex.Unlock()
				c.Error = errors.Errorf("Cannot make new log; canceling heartbeat. Please create a new client. Error: %v", err)
				c.heartbeatCancel()
				return
			}

			if err := c.deliverLog(ctx); err != nil {
				c.tickerMutex.Lock()
				newDuration := c.tickerDuration / BackoffGranularity
				if newDuration == 1 {
					newDuration++
				} else {
					newDuration = time.Duration(math.Pow(float64(newDuration), 2))
				}
				newDuration *= time.Second
				if newDuration > BackoffCap {
					newDuration = BackoffCap
				}
				c.tickerMutex.Unlock()

				c.ourLogger.Errorf("Received error while delivering log, increasing time until next delivery to %v: %v", newDuration, err)
				c.adjustTicker(newDuration)
				continue
			} else {
				if c.tickerDuration != DefaultTickInterval {
					c.ourLogger.Infof("Delivery succeeded; resetting to default interval %v", DefaultTickInterval)
					c.adjustTicker(DefaultTickInterval)
				}
			}
		}
	}
}

func (c *Client) deliverLog(ctx context.Context) error {
	return filepath.Walk(c.logDir, func(p string, fi os.FileInfo, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return err
		}

		if p == c.logDir {
			return nil
		}

		if fi.IsDir() {
			// we do not descend.
			c.ourLogger.Debugf("Directory (%v) found in log dir, skipping", p)
			return filepath.SkipDir
		}

		if fi.Mode()&os.ModeType != 0 {
			// irregular file, eject
			c.ourLogger.Debugf("Irregular file (%v) found in log dir, skipping", p)
			return nil
		}

		c.logLock.Lock()
		if c.logFile != nil {
			if path.Base(p) == path.Base(c.logFile.Name()) {
				c.logLock.Unlock()
				// we don't want to operate on the open file
				return nil
			}
			c.logLock.Unlock()
		}

		if fi.Size() == 0 {
			// it's not the current file so it's probably garbage
			if err := os.Remove(p); err != nil {
				c.ourLogger.Errorf("Could not remove empty file %q: %v", p, err)
				return nil
			}

			return nil
		}

		if err := c.sendLog(ctx, p); err != nil {
			c.ourLogger.Errorf("Could not deliver log bundle %v; will retry at next heartbeat. Error: %v", p, err)
			return err
		}

		return nil
	})
}

func (c *Client) sendLog(ctx context.Context, p string) (retErr error) {
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()
	defer func() {
		// if we cannot send or otherwise operate, do not remove the file, just log
		// the error and return; we will retry on the next heartbeat.
		if retErr != nil {
			c.ourLogger.Error(retErr)
			return
		}

		// remove the file if we received no error so it can't be re-delivered.
		if err := os.Remove(p); err != nil {
			c.ourLogger.Error(err)
		}
	}()

	client, err := c.logPipeClient.ReceiveLogs(ctx)
	if err != nil {
		return err
	}

	buf := make([]byte, 2*1024*1024)

	for {
		n, err := f.Read(buf)
		if err != nil {
			if err == io.EOF {
				if _, err := client.CloseAndRecv(); err != nil && err != io.EOF {
					return err
				}

				return nil
			}
		}

		if err := client.Send(&LogData{Data: buf[:n]}); err != nil {
			return err
		}
	}
}

func (c *Client) makeLog(takeLock bool) error {
	if takeLock {
		c.logLock.Lock()
		defer c.logLock.Unlock()
	}

	f, err := ioutil.TempFile(c.logDir, "")
	if err != nil {
		return err
	}

	if c.logFile != nil {
		c.logFile.Close()
	}

	c.logFile = f
	return nil
}

func (c *Client) Write(e *logrus.Entry) error {
	c.errorMutex.RLock()
	if c.Error != nil {
		defer c.errorMutex.RUnlock()
		return c.Error
	}
	c.errorMutex.RUnlock()

	f := &types.Struct{Fields: map[string]*types.Value{}}
	for key, value := range e.Data {
		var v string
		switch value.(type) {
		case string:
			v = value.(string)
		default:
			v = fmt.Sprintf("%v", value)
		}

		f.Fields[key] = &types.Value{Kind: &types.Value_StringValue{StringValue: v}}
	}

	t, err := types.TimestampProto(e.Time)
	if err != nil {
		return err
	}

	eOut := &Entry{
		Level:   int64(e.Level),
		Fields:  f,
		Message: e.Message,
		At:      t,
		Service: c.service,
	}

	buf, err := eOut.Marshal()
	if err != nil {
		return err
	}

	c.logLock.Lock()
	defer c.logLock.Unlock()

	if c.logFile == nil {
		if err := c.makeLog(true); err != nil {
			return err
		}
	}

	if err := binary.Write(c.logFile, binary.BigEndian, int64(len(buf))); err != nil {
		return err
	}

	if _, err := c.logFile.Write(buf); err != nil {
		return err
	}

	return nil
}
