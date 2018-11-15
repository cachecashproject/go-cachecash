// Modified from a gorilla example.  XXX: TMP: Intended for temporary development use; remove me.
// N.B.: Run this with `go run ./client/temphardwired.go`.

// +build ignore

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"
	cachecash "github.com/kelleyk/go-cachecash"
	"github.com/kelleyk/go-cachecash/ccmsg"
	"github.com/kelleyk/go-cachecash/common"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var clientPublicKey ed25519.PublicKey

func prepareMessage(sequenceNo uint64) []byte {
	msg := &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(clientPublicKey),
		Path:            "/foo/bar",
		BlockIdx: []uint64{
			(sequenceNo * 4) + 0,
			(sequenceNo * 4) + 1,
			(sequenceNo * 4) + 2,
			(sequenceNo * 4) + 3,
		},
		SequenceNo: sequenceNo,
	}

	msgData, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return msgData
}

// XXX: Not sure what the unique identifier for a cache is just yet; it should not be the network address, so that we
// can handle address migrations.  (Or maybe it's okay to ignore this from the client's perspective?)
type cacheID string

type client struct {
	l *logrus.Logger

	cacheConns map[cacheID]*cacheConn
}

// XXX: This should be merged with `wire/connection.go`; there are a lot of similarities.
type cacheConn struct {
	l *logrus.Logger

	c *client

	remoteAddr string

	outboundCh chan []byte
	// closed on disconnect
	inboundCh chan []byte

	// XXX: This is probably not how we actually want to pass this data around.
	bundle         *ccmsg.TicketBundle
	cacheBundleIdx int
}

func (cc *cacheConn) Send(msg proto.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "failed to marshal request to cache")
	}
	cc.outboundCh <- data
	return nil
}

func (cc *cacheConn) run(ctx context.Context) {
	u := url.URL{Scheme: "ws", Host: cc.remoteAddr, Path: "/api/v0/client"}
	log.Printf("connecting to cache: %s", u.String())

	co, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer co.Close()

	log.Printf("connected to cache: %s", u.String())
	go cc.runRx(ctx, co)

	go func() {
		for data := range cc.inboundCh {
			log.Printf("data arrived from cache\n")
			msg := &ccmsg.ClientCacheResponse{}
			if err := proto.Unmarshal(data, msg); err != nil {
				panic(err)
			}
			if err := cc.handleCacheMessage(msg); err != nil {
				panic(err)
			}
		}
	}()

	for {
		select {
		case data := <-cc.outboundCh:
			fmt.Printf("sending message to cache\n")
			err := co.WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				log.Println("cache write error:", err)
				return
			}
		}
	}
}

func (cc *cacheConn) runRx(ctx context.Context, co *websocket.Conn) {
	cc.l.Info("(TMP) entering runRx")

	defer cc.cleanupRx()

	for {
		msgType, data, err := co.ReadMessage()
		if err != nil {
			switch {
			case common.IsWebSocketClose(err, websocket.CloseNormalClosure, websocket.CloseGoingAway):
				cc.l.Info("connection closed gracefully")
			case common.IsNetErrClosing(err):
				// If we got this particular error, another goroutine must have closed our connection for us.
				cc.l.Warn("connection appears to have been uncleanly closed from our side")
			default:
				// The other side has uncleanly closed the connection; we should exit without trying to send a graceful
				// close message.
				cc.l.WithError(err).Error("failed to read WebSocket message")
			}

			return
		}

		switch msgType {
		case websocket.BinaryMessage:
			cc.l.Info("got message; sending to inboundCh")
			cc.inboundCh <- data
		case websocket.CloseMessage:
			// TODO: handle me; close gracefully
			cc.l.Error("recv websock close message, but graceful close is not implemented")
			return
		default:
			cc.l.Error("received a WebSocket message, but it is not a binary message")
		}
	}
}

func (cc *cacheConn) cleanupRx() {
	// close(cc.finishedCh)
	close(cc.inboundCh)
}

func (cc *cacheConn) handleCacheMessage(resp *ccmsg.ClientCacheResponse) error {
	switch resp.Msg.(type) {
	case *ccmsg.ClientCacheResponse_DataResponse:
		// TODO: Validate message before sending L1 ticket.

		log.Printf("got cache message: data response\n")
		msg, err := cc.bundle.BuildClientCacheRequest(cc.bundle.TicketL1[cc.cacheBundleIdx])
		if err != nil {
			return errors.Wrap(err, "failed to generate L1 message")
		}
		if err := cc.Send(msg); err != nil {
			return errors.Wrap(err, "failed to send L1 message")
		}
	case *ccmsg.ClientCacheResponse_L1Response:
		log.Printf("got cache message: L1 response\n")

		// TODO: Need synchronization between all related cache-client requests before we can solve the colocation
		// puzzle.
		/*
			bundle := cc.bundle

			// XXX: This section cribbed from `integration_Test.go`.

			// Solve colocation puzzle.
			pi := bundle.Remainder.PuzzleInfo
			secret, _, err := colocationpuzzle.Solve(colocationpuzzle.Parameters{
				Rounds:      pi.Rounds,
				StartOffset: uint32(pi.StartOffset),
				StartRange:  uint32(pi.StartRange),
			}, singleEncryptedBlocks, pi.Goal)
			if err != nil {
				return err
			}

			// Decrypt L2 ticket.
			ticketL2, err := common.DecryptTicketL2(secret, bundle.EncryptedTicketL2)
			if err != nil {
				return err
			}

			// Give L2 tickets to caches.
			for _, cache := range caches {
				msg, err := bundle.BuildClientCacheRequest(&ccmsg.TicketL2Info{
					EncryptedTicketL2: bundle.EncryptedTicketL2,
					PuzzleSecret:      secret,
				})


			msg, err := cc.bundle.BuildClientCacheRequest(cc.bundle.TicketL2[cc.cacheBundleIdx])
			if err != nil {
				return errors.Wrap(err, "failed to generate L2 message")
			}
			if err := cc.Send(msg); err != nil {
				return errors.Wrap(err, "failed to send L2 message")
			}
		*/
	case *ccmsg.ClientCacheResponse_L2Response:
		log.Printf("got cache message: L2 response\n")
		// ... done!
	default:
		log.Printf("got cache message: UNKNOWN TYPE\n")
	}
	return nil
}

func (c *client) sendProviderMessage(req *ccmsg.ContentRequest) error {
	// TODO: ...
	return nil
}

func (c *client) handleProviderMessage(resp *ccmsg.ContentResponse) error {
	if resp.Error != nil {
		return fmt.Errorf("got error from provider: %v", resp.Error.Message)
	}

	bundle := resp.Bundle
	if bundle == nil {
		return errors.New("got nil bundle from provider")
	}

	// TODO: Validate response.  TicketRequest and TicketL1 and CacheInfo must have same length.

	for i := 0; i < len(bundle.TicketRequest); i++ {
		// TEMP: see comment on cacheID type
		addr := bundle.CacheInfo[i].Addr.ConnectionString()
		if addr == "" {
			panic("no connection string for cache")
		}
		cacheID := (cacheID)(addr)

		cc, ok := c.cacheConns[cacheID]
		if !ok {
			cc = &cacheConn{
				l:              c.l,
				c:              c,
				remoteAddr:     addr,
				bundle:         bundle,
				cacheBundleIdx: i,
				outboundCh:     make(chan []byte),
				inboundCh:      make(chan []byte),
			}
			ctx := context.Background() // TEMP: should be something we can cancel
			go cc.run(ctx)
			c.cacheConns[cacheID] = cc
		}

		cacheReq, err := bundle.BuildClientCacheRequest(bundle.TicketRequest[i])
		if err != nil {
			return errors.Wrap(err, "failed to generate request to cache")
		}
		if err := cc.Send(cacheReq); err != nil {
			return errors.Wrap(err, "failed to send message to cache")
		}
	}

	return nil
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/api/v0/ws"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	maxRequests := 1

	// ---- stuff related to CacheCash specifically ----

	cl := &client{
		l:          logrus.New(),
		cacheConns: make(map[cacheID]*cacheConn),
	}

	// // Maps sequence number to the request that we sent.
	// contentRequests := make(map[int]*ccmsg.ContentRequest)

	// Create a client keypair.
	var clientPrivateKey ed25519.PrivateKey
	clientPublicKey, clientPrivateKey, err = ed25519.GenerateKey(nil)
	if err != nil {
		// return errors.Wrap(err, "failed to generate client keypair")
		panic("failed to generate client keypair")
	}
	_ = clientPrivateKey

	// ---- end ----

	go func() {
		defer close(done)
		for {
			_, msgData, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			bundle := &ccmsg.ContentResponse{}
			if err := proto.Unmarshal(msgData, bundle); err != nil {
				log.Println("unmarshal:", err)
				return
			}

			cl.handleProviderMessage(bundle)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	requestNo := 0
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			// XXX: We deliberately wait until we *would* sned the next message in order to allow for responses to
			// arrive.  This is horrible.
			requestNo++
			if requestNo > maxRequests {
				// XXX: This is terrible; exit by faking SIGINT.
				// interrupt <- os.Interrupt
				continue
			}
			err := c.WriteMessage(websocket.BinaryMessage, prepareMessage(uint64(requestNo-1)))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
