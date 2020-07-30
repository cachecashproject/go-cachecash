package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cachecashproject/go-cachecash/common"
	"github.com/cachecashproject/go-cachecash/kv"
	"github.com/cachecashproject/go-cachecash/kv/ratelimit"
	"github.com/cachecashproject/go-cachecash/log"
	es "github.com/elastic/go-elasticsearch/v7"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	_ "github.com/lib/pq" // postgres driver
)

const (
	// filename of process list file (lives in spooldir)
	processListFileNameSuffix = "process_list.json"

	// DefaultLogSize is the maximum size bundle we will accept. If unaccepted, it
	// will return ErrBundleTooLarge to the grpc client.
	DefaultLogSize = 10 * 1024 * 1024
)

// FileMeta is a simple struct for passing ip info with filenames... Private; mostly public for json serialization.
type FileMeta struct {
	PubKey []byte
	Name   string
	IPAddr string
}

// Config is the configuration of the server.
type Config struct {
	IndexName     string
	ListenAddress string
	SpoolDir      string
	MaxLogSize    uint64
	ElasticSearch es.Config
	KVMember      string
	Logger        *logrus.Logger

	RateLimiting    bool
	RateLimitConfig ratelimit.Config

	Processor func(file *FileMeta) error
}

func (c *Config) processListFileName() string {
	return filepath.Join(c.SpoolDir, processListFileNameSuffix)
}

// LogPipe pipes data from protobuf-generated logs (see log/client.go for the
// client) to elasticsearch in a structured fashion. Logs are delivered as
// bundles, which are signed and tagged with a public key that corresponds to
// the client. Each bundle is stored locally and delivered to ES from disk.
type LogPipe struct {
	config Config

	server    *grpc.Server
	listener  net.Listener
	connMutex sync.RWMutex // mutex for controlling server state

	esClient *es.Client
	kvClient *kv.Client

	ratelimiter *ratelimit.RateLimiter

	processList      []*FileMeta
	processListMutex sync.Mutex
	processContext   context.Context
	processCancel    context.CancelFunc
}

// NewLogPipe creates a LogPipe server.
func NewLogPipe(config Config) (*LogPipe, error) {
	esClient, err := es.NewClient(config.ElasticSearch)
	if err != nil {
		return nil, err
	}

	kv := kv.NewClient(config.KVMember, kv.NewDBDriver(config.Logger))

	var processList []*FileMeta

	f, err := os.Open(config.processListFileName())
	if err != nil {
		processList = []*FileMeta{}
	} else {
		if err := json.NewDecoder(f).Decode(&processList); err != nil {
			processList = []*FileMeta{}
		}
	}

	var ratelimiter *ratelimit.RateLimiter

	if config.RateLimiting {
		config.RateLimitConfig.Logger = config.Logger
		ratelimiter = ratelimit.NewRateLimiter(config.RateLimitConfig, kv)
	}

	lp := &LogPipe{
		processList: processList,

		kvClient:    kv,
		esClient:    esClient,
		ratelimiter: ratelimiter,
		config:      config,
	}
	return lp, nil
}

// ListenAddr reports the listener address, especially for when binding to :0.
// returns an empty string if the listener is empty.
func (lp *LogPipe) ListenAddr() string {
	lp.connMutex.Lock()
	defer lp.connMutex.Unlock()
	if lp.listener != nil {
		return lp.listener.Addr().String()
	}

	return ""
}

func (lp *LogPipe) persistProcessList() error {
	lp.processListMutex.Lock()
	defer lp.processListMutex.Unlock()
	// XXX we do this in the spool dir so that os.Rename() won't complain about hard
	// links being cross-device.
	// the whole point here is to avoid a partial write to the process list.
	f, err := ioutil.TempFile(lp.config.SpoolDir, "process_list.tmp.")
	if err != nil {
		return err
	}

	if err := json.NewEncoder(f).Encode(lp.processList); err != nil {
		f.Close()
		return err
	}

	if err := f.Sync(); err != nil {
		f.Close() // pray
		return err
	}

	f.Close()

	dir, err := os.Open(lp.config.SpoolDir)
	if err != nil {
		return err
	}
	defer func() {
		if err := dir.Sync(); err != nil {
			lp.config.Logger.Errorf("error during directory sync: %v", err)
		}

		dir.Close()
	}()

	return os.Rename(f.Name(), lp.config.processListFileName())
}

// Close closes the server. If a timeout is given, it will sleep for that long
// between shutting down grpc and closing the listener, allowing for cleanup to
// potentially occur.
func (lp *LogPipe) Close(timeout time.Duration) error {
	lp.connMutex.RLock()
	defer lp.connMutex.RUnlock()
	lp.processCancel()
	if err := lp.persistProcessList(); err != nil {
		logrus.Errorf("Could not persist process list: %v", err)
		// XXX not returning here so the server can stop -- this will lose logs
	}

	// FIXME wait for processors to complete
	lp.server.GracefulStop()
	lp.config.Logger.Warnf("Sleeping for %v to allow processes to complete", timeout)
	time.Sleep(timeout)

	if lp.listener != nil {
		return lp.listener.Close()
	}

	return nil
}

// Boot boots the service. It returns error if it cannot, or if there was an
// error that occurred while serving. This function will block until the server
// is stopped. It will close the `ready` channel you pass when ready to serve.
func (lp *LogPipe) Boot(ready chan struct{}, db *sql.DB) error {
	lp.connMutex.Lock()

	lp.server = common.NewDBGRPCServer(db)
	log.RegisterLogPipeServer(lp.server, lp)
	lp.processContext, lp.processCancel = context.WithCancel(context.Background())

	var err error
	lp.listener, err = net.Listen("tcp", lp.config.ListenAddress)
	if err != nil {
		lp.connMutex.Unlock()
		return err
	}

	lp.connMutex.Unlock()
	close(ready)

	go lp.process()
	return lp.server.Serve(lp.listener)
}
