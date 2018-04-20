// Command kobject opens a Linux kobject userspace events listener and prints
// all received events to the terminal.
package main

import (
	"fmt"
	"log"

	"github.com/mdlayher/kobject"
)

func main() {
	c, err := kobject.New()
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer c.Close()

	for {
		event, err := c.Receive()
		if err != nil {
			log.Fatalf("failed to receive: %v", err)
		}

		fmt.Println(event)
	}
}
