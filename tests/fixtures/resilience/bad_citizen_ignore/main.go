package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Catch signals and IGNORE them to force a kill
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Bad Citizen (Ignore) Started")

	go func() {
		for s := range sigs {
			fmt.Printf("Ignoring signal: %v\n", s)
		}
	}()

	// Block forever
	for {
		time.Sleep(1 * time.Second)
	}
}
