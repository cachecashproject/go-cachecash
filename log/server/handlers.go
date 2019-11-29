package server

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cachecashproject/go-cachecash/log"
	"golang.org/x/crypto/ed25519"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// ReceiveLogs receives the logs for processing. Spins out a goroutine to send to ES once received.
func (lp *LogPipe) ReceiveLogs(lf log.LogPipe_ReceiveLogsServer) (retErr error) {
	peer, ok := peer.FromContext(lf.Context())
	if !ok {
		return status.Errorf(codes.FailedPrecondition, "failed to get grpc peer from ctx")
	}

	incomingIP := peer.Addr.String()[:strings.LastIndex(peer.Addr.String(), ":")]
	if err := os.MkdirAll(filepath.Join(lp.config.SpoolDir, incomingIP), 0700); err != nil {
		return status.Errorf(codes.ResourceExhausted, err.Error())
	}

	pubkey, err := lf.Recv()
	if err != nil {
		return err
	}

	if len(pubkey.PubKey) != 32 {
		return status.Errorf(codes.InvalidArgument, "pubkey is the wrong size")
	}

	dataRateKey := strings.Join([]string{"rate-limit", "data", incomingIP}, "/")

	tf, err := os.Create(filepath.Join(lp.config.SpoolDir, incomingIP, fmt.Sprintf("%d", time.Now().Unix())))
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, err.Error())
	}
	defer func() {
		tf.Close() // may already be closed, don't check the error here
		if retErr != nil {
			os.Remove(tf.Name()) // we're depending on the client to re-send the log bundle
		}
	}()

	var size uint64

	for {
		select {
		case <-lp.processContext.Done():
			return lp.processContext.Err()
		case <-lf.Context().Done():
			return lf.Context().Err() // return an error so the file gets cleaned up
		default:
		}

		data, err := lf.Recv()
		if err != nil {
			if err != io.EOF {
				return status.Errorf(codes.FailedPrecondition, err.Error())
			}

			if err := tf.Close(); err != nil {
				return status.Errorf(codes.Unknown, err.Error())
			}

			// since there was no error, this file will not be cleaned up in the
			// defer above. Instead, process will do that once its done working.

			lp.processListMutex.Lock()
			defer lp.processListMutex.Unlock()
			lp.processList = append(lp.processList, &FileMeta{PubKey: pubkey.PubKey, Name: tf.Name(), IPAddr: incomingIP})

			return nil
		}

		dataSize := int64(len(data.Data))
		size += uint64(dataSize)

		if lp.ratelimiter != nil {
			if err := lp.ratelimiter.RateLimit(lf.Context(), dataRateKey, dataSize); err != nil {
				return status.Errorf(codes.Unavailable, err.Error())
			}
		}

		if size > lp.config.MaxLogSize {
			return status.Errorf(codes.InvalidArgument, log.ErrBundleTooLarge.Error())
		}

		if !ed25519.Verify(pubkey.PubKey, data.Data, data.Signature) {
			return status.Errorf(codes.InvalidArgument, "log bundle was not signed properly")
		}

		if _, err := tf.Write(data.Data); err != nil {
			return status.Errorf(codes.ResourceExhausted, err.Error())
		}
	}
}
