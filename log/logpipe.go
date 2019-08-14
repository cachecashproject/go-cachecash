package log

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	es "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/peer"
)

// DefaultLogSize is the maximum size bundle we will accept. If unaccepted, it
// will return ErrBundleTooLarge to the grpc client.
const DefaultLogSize = 100 * 1024 * 1024

// a simple channel pair for passing ip info with filenames
type filepair struct {
	name   string
	ipaddr string
}

// Pipe is the name of the receiving service that proxies protos into ES.
type Pipe struct {
	IndexName       string
	SpoolDir        string
	MaxSpoolDirSize int64
	MaxLogSize      int64
	MaxConnections  int
	Processors      int

	processorCancel context.CancelFunc
	connCountMutex  sync.Mutex
	connCount       int
	esClient        *es.Client
	fileQueue       chan filepair
}

// NewPipeServer creates a new GRPC logpipe server.
func NewPipeServer(indexName string, processors int, spoolDir string, esConfig es.Config) (*Pipe, error) {
	es, err := es.NewClient(esConfig)
	if err != nil {
		return nil, err
	}

	p := &Pipe{
		IndexName:  indexName,
		Processors: processors,
		SpoolDir:   spoolDir,
		MaxLogSize: DefaultLogSize,

		esClient:  es,
		fileQueue: make(chan filepair, processors), // FIXME This was a dumb idea
	}
	p.startProcessors()
	return p, nil
}

func (p *Pipe) startProcessors() {
	ctx, cancel := context.WithCancel(context.Background())
	p.processorCancel = cancel

	for i := 0; i < p.Processors; i++ {
		go p.process(ctx)
	}
}

func (p *Pipe) connDel() {
	if p.MaxConnections != 0 {
		p.connCountMutex.Lock()
		defer p.connCountMutex.Unlock()
		p.connCount--
	}
}

func (p *Pipe) connAdd() error {
	if p.MaxConnections != 0 {
		p.connCountMutex.Lock()
		defer p.connCountMutex.Unlock()
		p.connCount++
		if p.connCount > p.MaxConnections {
			return ErrTooManyConnections
		}
	}

	return nil
}

func (p *Pipe) spoolDirCheck() error {
	if p.MaxSpoolDirSize == 0 {
		return nil
	}

	var total int64
	err := filepath.Walk(p.SpoolDir, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			logrus.Errorf("Error encountered stating file %q: %v", p, err)
			return nil
		}

		if fi.IsDir() {
			// these are invalid but we really don't care and throwing an error here has consequences.
			logrus.Debugf("Invalid dir %q found while scanning spool dir", p)
			return filepath.SkipDir
		}

		if fi.Mode()&os.ModeType != 0 {
			// irregular file, eject
			logrus.Debugf("Irregular file (%v) found in log dir, skipping", p)
			return nil
		}

		total += fi.Size()
		return nil
	})
	if err != nil {
		// in this case, walk actually failed for some reasons since we don't
		// return any errors from inside. we want to return ErrSpoolFull here
		// because we don't want to leak this info to the client, but we log it
		// anyway.
		logrus.Errorf("Error walking spool dir: %v", err)
		return ErrSpoolFull
	}

	// 5% of the storage is reserved for sanity. then we make sure the log will fit.
	if float64(total) >= ((float64(p.MaxSpoolDirSize) * 0.95) - float64(p.MaxLogSize)) {
		return ErrSpoolFull
	}

	return nil
}

// ReceiveLogs receives the logs for processing. Spins out a goroutine to send to ES once received.
func (p *Pipe) ReceiveLogs(lf LogPipe_ReceiveLogsServer) (retErr error) {
	if err := p.connAdd(); err != nil {
		return grpcFailedError(err)
	}
	defer p.connDel()

	if err := p.spoolDirCheck(); err != nil {
		return grpcFailedError(err)
	}

	incomingIP := "unknown"

	peer, ok := peer.FromContext(lf.Context())
	if !ok {
		return grpcFailedError(errors.New("failed to get grpc peer from ctx"))
	}

	if ok {
		incomingIP = peer.Addr.String()
	}

	tf, err := ioutil.TempFile(p.SpoolDir, "")
	if err != nil {
		return grpcFailedError(err)
	}
	defer func() {
		tf.Close() // may already be closed, don't check the error here
		if retErr != nil {
			os.Remove(tf.Name())
		}
	}()

	var size int64

	for {
		select {
		case <-lf.Context().Done():
			return lf.Context().Err() // return an error so the file gets cleaned up
		default:
		}

		data, err := lf.Recv()
		if err != nil {
			if err != io.EOF {
				return grpcFailedError(err)
			}

			if err := tf.Close(); err != nil {
				return grpcFailedError(err)
			}

			// since there was no error, this file will not be cleaned up in the
			// defer above. Instead, process will do that once its done working.

			p.fileQueue <- filepair{name: tf.Name(), ipaddr: incomingIP}
			return nil
		}

		size += int64(len(data.Data))
		if size > p.MaxLogSize {
			return grpcFailedError(ErrBundleTooLarge)
		}

		if _, err := tf.Write(data.Data); err != nil {
			return grpcFailedError(err)
		}
	}
}

func (p *Pipe) processFile(file filepair) {
	r, err := NewReader(file.name)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer r.Close()
	defer os.Remove(file.name)

	for {
		e, err := r.NextProto()
		if err != nil {
			if err != io.EOF {
				logrus.Error(err)
			}
			return
		}

		ts, err := types.TimestampFromProto(e.At)
		if err != nil {
			logrus.Error(err)
			return
		}

		fields := map[string]string{}
		for key, value := range e.Fields.Fields {
			fields[key] = value.GetStringValue()
		}

		fields["peer_ip_address"] = file.ipaddr[:strings.LastIndex(file.ipaddr, ":")]
		fields["log_time"] = ts.Format(time.RFC3339)
		fields["service"] = e.Service // special case for our service tags
		fields["log_level"] = logrus.Level(e.Level).String()
		fields["message"] = e.Message

		buf, err := json.Marshal(fields)
		if err != nil {
			logrus.Error(err)
			return
		}

		ir := esapi.IndexRequest{
			Index: p.IndexName,
			Body:  bytes.NewBuffer(buf),
		}

		res, err := ir.Do(context.Background(), p.esClient)
		if err != nil {
			logrus.Error(err)
			return
		}
		res.Body.Close()

		if res.IsError() {
			logrus.Errorf("Error indexing document: %v", res.Status())
		}
	}
}

func (p *Pipe) process(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case res := <-p.fileQueue:
			p.processFile(res)
		}
	}
}
