// Command kobject opens a Linux kobject userspace events listener and prints
// all received events to the terminal.
package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/mdlayher/kobject"
)

func main() {
	var (
		tFlag = flag.Duration("t", 0*time.Second, "the amount of time to wait between events before timing out (default: forever)")
		rFlag = flag.Int("r", 0, "the size of the read buffer for kobject events (0 for default size)")
	)
	flag.Parse()

	c, err := kobject.New()
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer c.Close()

	if *rFlag != 0 {
		if err = c.SetReadBuffer(*rFlag); err != nil {
			log.Fatalf("failed to set read buffer: %v", err)
		}
	}

	for {
		if *tFlag > 0 {
			if err := c.SetDeadline(time.Now().Add(*tFlag)); err != nil {
				log.Fatalf("failed to set deadline: %v", err)
			}
		}

		event, err := c.Receive()
		if err != nil {
			log.Fatalf("failed to receive: %v", err)
		}

		fmt.Println(event)
	}
}
