// Modified from a gorilla example.
// XXX: TMP: Intended for temporary development use; remove me.
// This "simple" variant only communicates with the provider.  It does not try to do anything with the caches.

// +build ignore

package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"
	cachecash "github.com/kelleyk/go-cachecash"
	"github.com/kelleyk/go-cachecash/ccmsg"
	"golang.org/x/crypto/ed25519"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var clientPublicKey ed25519.PublicKey

func prepareMessage() []byte {
	msg := &ccmsg.ContentRequest{
		ClientPublicKey: cachecash.PublicKeyMessage(clientPublicKey),
		Path:            "/foo/bar",
		BlockIdx: []uint64{
			0,
			1,
			2,
			3,
		},
		SequenceNo: 42,
	}

	msgData, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return msgData
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

	maxRequests := 2

	// ---- stuff related to CacheCash specifically ----

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

			log.Printf("recv bundle: %v", bundle)
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
				interrupt <- os.Interrupt
				continue
			}
			err := c.WriteMessage(websocket.BinaryMessage, prepareMessage())
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
