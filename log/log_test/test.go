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
	c, err := log.NewClient("localhost:9005", "monkeys", "/tmp/1234")
	if err != nil {
		panic(err)
	}

	logrus.AddHook(log.NewHook(c))
	logrus.SetOutput(ioutil.Discard)

	after := time.After(10 * time.Second)
	var i int
	for ; true; i++ {
		select {
		case <-after:
			goto end
		default:
			logrus.WithField("count", i).Info("this is a message")
		}
	}

end:
	time.Sleep(2 * time.Second)
	c.Close()
	time.Sleep(2 * time.Second)
	fmt.Println(i)
}
