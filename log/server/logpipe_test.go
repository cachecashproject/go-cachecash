// +build external_test

package server

import (
	"os"
	"testing"
	"time"

	"github.com/cachecashproject/go-cachecash/log"
	"github.com/stretchr/testify/assert"
)

func Test01LogPipeBasic(t *testing.T) {
	lr := &logRecorder{}

	config, err := makeConfig(lr.recordLogs)
	assert.Nil(t, err)
	defer os.RemoveAll(config.SpoolDir)
	wipeTables(config)

	lp, errChan := readyServer(t, config)
	defer func() {
		lp.Close(time.Second)
		assert.Nil(t, <-errChan)
	}()

	client, logger, dir, err := makeClient(lp, "")
	defer os.RemoveAll(dir)
	defer client.Close()

	logger.Info("hi")
	time.Sleep(3 * time.Second)

	assert.Equal(t, 1, lr.Len())
	lr.checkEntries(t, func(t *testing.T, e *log.Entry) {
		assert.Equal(t, e.Message, "hi")
	})
}

func TestLogPipeSizeLimit(t *testing.T) {
	lr := &logRecorder{}

	config, err := makeConfig(lr.recordLogs)
	assert.Nil(t, err)
	defer os.RemoveAll(config.SpoolDir)
	wipeTables(config)

	lp, errChan := readyServer(t, config)
	defer func() {
		lp.Close(time.Second)
		assert.Nil(t, <-errChan)
	}()

	client, logger, dir, err := makeClient(lp, "")
	assert.Nil(t, err)
	defer os.RemoveAll(dir)

	// this should be too large for the pre-configured payload max
	for i := 0; i < 50; i++ {
		logger.Info("hi")
	}

	time.Sleep(2 * time.Second)

	assert.Equal(t, 0, lr.Len())
	lp.Close(time.Second)
	assert.Nil(t, <-errChan)
	assert.Nil(t, client.Close())
	// change the max ;)

	lp.config.MaxLogSize = 1024 * 1024
	ready := make(chan struct{})
	go func() {
		errChan <- lp.Boot(ready)
	}()
	defer lp.Close(time.Second)
	<-ready
	client, logger, dir, err = makeClient(lp, dir)
	assert.Nil(t, err)
	defer client.Close()
	time.Sleep(5 * time.Second)
	assert.Equal(t, 50, lr.Len())
}

func TestLogPipeDataRateLimit(t *testing.T) {
	lr := &logRecorder{}

	config, err := makeConfig(lr.recordLogs)
	assert.Nil(t, err)

	// these values make/break the test. tune carefully with DEBUG=1.
	config.MaxLogSize = 2048
	config.RateLimiting = true
	config.RateLimitConfig.Cap = 2048
	config.RateLimitConfig.RefreshAmount = 32
	config.RateLimitConfig.RefreshInterval = time.Second
	defer os.RemoveAll(config.SpoolDir)
	wipeTables(config)

	lp, errChan := readyServer(t, config)
	defer func() {
		lp.Close(time.Second)
		assert.Nil(t, <-errChan)
	}()

	client, logger, dir, err := makeClient(lp, "")
	assert.Nil(t, err)
	defer client.Close()
	defer os.RemoveAll(dir)

	// this should be enough to eat away at the tokens
	for i := 0; i < 20; i++ {
		logger.Info("hi")
	}
	time.Sleep(250 * time.Millisecond)

	// this should over-fill the tokens
	for i := 0; i < 20; i++ {
		logger.Info("hi")
	}
	time.Sleep(250 * time.Millisecond)

	// this should over-fill the tokens
	for i := 0; i < 20; i++ {
		logger.Info("hi")
	}

	after := time.After(5 * time.Second)
	for {
		select {
		case <-after:
			assert.Fail(t, "failed to get all the data")
			break
		default:
		}

		if lr.Len() == 60 {
			assert.Equal(t, 60, lr.Len())
			break
		}
	}
}
