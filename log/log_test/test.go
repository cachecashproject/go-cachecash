// this is a basic log shipper for testing logpiped. Not intended for real use.
package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/cachecashproject/go-cachecash/log"
	"github.com/sirupsen/logrus"
)

func main() {
	c, err := log.NewClient("localhost:9005", "monkeys", "/tmp/1234", true, true, log.DefaultConfig())
	if err != nil {
		panic(err)
	}

	l := logrus.New()
	l.SetOutput(ioutil.Discard)
	l.AddHook(log.NewHook(c))

	after := time.After(10 * time.Second)
	var i int
	for ; true; i++ {
		select {
		case <-after:
			goto end
		default:
			l.WithField("count", i).Info("this is a message")
		}
	}

end:
	time.Sleep(2 * time.Second)
	c.Close()
	time.Sleep(2 * time.Second)
	fmt.Println(i)
}
