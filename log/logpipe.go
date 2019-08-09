package log

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	es "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Pipe is the name of the receiving service that proxies protos into ES.
type Pipe struct {
	spoolDir   string
	maxLogSize int64

	esClient *es.Client
}

// NewPipeServer creates a new GRPC logpipe server.
func NewPipeServer(spoolDir string, maxLogSize int64) (*Pipe, error) {
	es, err := es.NewDefaultClient()
	if err != nil {
		return nil, err
	}

	return &Pipe{spoolDir: spoolDir, maxLogSize: maxLogSize, esClient: es}, nil
}

func failedError(err error) error {
	return status.Errorf(codes.FailedPrecondition, "%v", err)
}

// ReceiveLogs receives the logs for processing. Spins out a goroutine to send to ES once received.
func (p *Pipe) ReceiveLogs(lf LogPipe_ReceiveLogsServer) (retErr error) {
	tf, err := ioutil.TempFile(p.spoolDir, "")
	if err != nil {
		return failedError(err)
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
				return failedError(err)
			}

			if err := tf.Close(); err != nil {
				return failedError(err)
			}

			// since there was no error, this file will not be cleaned up in the
			// defer above. Instead, process will do that once its done working.

			go p.process(tf.Name())
			return nil
		}

		size += int64(len(data.Data))
		if size > p.maxLogSize {
			return failedError(errors.New("payload too large"))
		}

		n, err := tf.Write(data.Data)
		if err != nil {
			return failedError(err)
		}

		if n != len(data.Data) {
			return failedError(errors.New("could not complete write to disk"))
		}
	}
}

func (p *Pipe) process(filename string) {
	r, err := NewReader(filename)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer r.Close()
	defer os.Remove(filename)

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

		fields["log_time"] = fmt.Sprintf("%v", ts)
		fields["service"] = e.Service // special case for our service tags
		fields["log_level"] = logrus.Level(e.Level).String()
		fields["message"] = e.Message

		buf, err := json.Marshal(fields)
		if err != nil {
			logrus.Error(err)
			return
		}

		ir := esapi.IndexRequest{
			Index: "test",
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
