package log

import (
	"errors"
	"io"
	"io/ioutil"
	"net"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// TestPipe is the name of the receiving service for testing the logger client
type TestPipe struct {
	// something to keep -race from flipping out
	Mutex sync.Mutex

	// if set, will raise the error on the next read.
	RaiseError error

	l    net.Listener
	s    *grpc.Server
	file string

	serveMutex sync.Mutex
}

// NewPipeServer creates a new GRPC logpipe server.
func NewTestPipeServer(file string) *TestPipe {
	return &TestPipe{file: file}
}

func (p *TestPipe) Serve(listenAddress string) error {
	p.serveMutex.Lock()

	p.s = grpc.NewServer()
	RegisterLogPipeServer(p.s, p)

	var err error
	p.l, err = net.Listen("tcp", listenAddress)
	if err != nil {
		p.serveMutex.Unlock()
		return err
	}

	p.serveMutex.Unlock()
	return p.s.Serve(p.l)
}

func (p *TestPipe) ListenAddress() string {
	for {
		p.serveMutex.Lock()
		if p.l != nil {
			p.serveMutex.Unlock()
			break
		}
		p.serveMutex.Unlock()
	}
	return p.l.Addr().String()
}

func (p *TestPipe) Close() {
	p.serveMutex.Lock()
	defer p.serveMutex.Unlock()

	if p.s != nil {
		p.s.GracefulStop()
	}

	if p.l != nil {
		p.l.Close()
	}
}

// ReceiveLogs receives the logs for processing. A large part of what this does
// is attempt to append to a single file in order from what it received; this
// way the whole corpus can be evaluated as a whole unit. There are no
// concurrency checks, so be aware of this when using it with your tests. If
// you set RaiseError, it will be raised after the last write you performed.
func (p *TestPipe) ReceiveLogs(lf LogPipe_ReceiveLogsServer) (retErr error) {
	tf, err := ioutil.TempFile("", "")
	if err != nil {
		return grpcFailedError(err)
	}
	defer func() {
		tf.Close()
		defer os.Remove(tf.Name())

		if retErr != nil { // if we had an error, discard the data -- it should be redelivered.
			logrus.Error(retErr)
			return
		}

		// append the temporary file we created for this run to the master file.
		f, err := os.OpenFile(p.file, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			logrus.Error(err)
			return
		}
		defer f.Close()

		tf, err := os.Open(tf.Name())
		if err != nil {
			logrus.Error(err)
			return
		}
		defer tf.Close()

		if _, err := io.Copy(f, tf); err != nil {
			logrus.Error(err)
			return
		}
	}()

	for {
		select {
		case <-lf.Context().Done():
			return grpcFailedError(lf.Context().Err())
		default:
		}

		data, err := lf.Recv()
		if err != nil {
			if err != io.EOF {
				return grpcFailedError(err)
			}

			return nil
		}

		n, err := tf.Write(data.Data)
		if err != nil {
			return grpcFailedError(err)
		}

		if n != len(data.Data) {
			return grpcFailedError(errors.New("could not complete write to disk"))
		}

		p.Mutex.Lock()
		if p.RaiseError != nil {
			defer p.Mutex.Unlock()
			return grpcFailedError(p.RaiseError)
		}
		p.Mutex.Unlock()
	}
}
