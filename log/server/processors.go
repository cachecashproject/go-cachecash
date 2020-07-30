package server

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"net"
	"os"
	"time"

	"github.com/cachecashproject/go-cachecash/log"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ErrFailedProcessing indicates a failure in processing that can be retried
var ErrFailedProcessing = errors.New("failed processing due to transient error; retrying")

func (lp *LogPipe) process() {
	if lp.config.Processor == nil {
		lp.config.Processor = lp.processFile
	}

	for {
		select {
		case <-lp.processContext.Done():
			logrus.Error(lp.processContext.Err())
			return
		default:
		}
		var process *FileMeta
		lp.processListMutex.Lock()
		if len(lp.processList) == 0 {
			time.Sleep(time.Second)
		} else {
			process = lp.processList[0]
			if len(lp.processList) == 1 {
				lp.processList = []*FileMeta{}
			} else {
				lp.processList = lp.processList[1:]
			}
		}
		lp.processListMutex.Unlock()

		if process != nil {
			if err := lp.config.Processor(process); err != nil {
				if err == ErrFailedProcessing {
					lp.config.Logger.Errorf("Failed during processing of log %v; retrying soon", process.Name)
					lp.processListMutex.Lock()
					lp.processList = append([]*FileMeta{process}, lp.processList...)
					lp.processListMutex.Unlock()
					time.Sleep(time.Second) // sleep a little in the hopes it will clear up
					continue
				}

				// FIXME don't just log... maybe stall processing on a lot of errors? re-fill the list?
				logrus.Error(err)
			}
		}
	}
}

func makeFields(e *log.Entry, file *FileMeta) (map[string]string, error) {
	ts, err := types.TimestampFromProto(e.At)
	if err != nil {
		return nil, err
	}

	fields := map[string]string{}
	for key, value := range e.Fields.Fields {
		fields[key] = value.GetStringValue()
	}

	fields["pubkey"] = hex.EncodeToString(file.PubKey)
	fields["peer_ip_address"] = file.IPAddr
	fields["log_time"] = ts.Format(time.RFC3339)
	fields["service"] = e.Service // special case for our service tags
	fields["log_level"] = logrus.Level(e.Level).String()
	fields["message"] = e.Message

	return fields, nil
}

func (lp *LogPipe) doIndex(ir esapi.IndexRequest) error {
	res, err := ir.Do(context.Background(), lp.esClient)
	if err != nil {
		return err
	}
	res.Body.Close()

	if res.IsError() {
		return errors.Errorf("Error indexing document: %v", res.Status())
	}

	return nil
}

func (lp *LogPipe) processFile(file *FileMeta) (retErr error) {
	r, err := log.NewReader(file.Name)
	if err != nil {
		return errors.Wrapf(err, "while reading in %v", file.Name)
	}
	defer r.Close()
	defer func() {
		if _, ok := retErr.(*net.OpError); ok {
			retErr = ErrFailedProcessing
			return
		}
		if retErr != nil {
			lp.config.Logger.Errorf("Error processing bundle, discarding contents: %v", retErr)
		}

		// it's ok if these error, but log the removal case in the situation where
		// we have a file with bad permissions.
		if err := os.Remove(file.Name); err != nil {
			lp.config.Logger.Warnf("Error received cleaning up log bundle: %v", err)
		}
	}()

	for {
		e, err := r.NextProto()
		if err != nil {
			if err != io.EOF {
				return err
			}
			return nil
		}

		fields, err := makeFields(e, file)
		if err != nil {
			return err
		}

		buf, err := json.Marshal(fields)
		if err != nil {
			return err
		}

		ir := esapi.IndexRequest{
			Index: lp.config.IndexName,
			Body:  bytes.NewBuffer(buf),
		}

		if err := lp.doIndex(ir); err != nil {
			return err
		}
	}
}
