package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Setup signal handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Bad Citizen (Slow) Started")

	go func() {
		sig := <-sigs
		fmt.Printf("\nReceived signal: %s\n", sig)
		fmt.Println("I heard you, but I am too busy sleeping...")
		// Sleep longer than the 5s grace period
		time.Sleep(10 * time.Second)
		fmt.Println("Fine, I'm done now (too late)")
		os.Exit(0)
	}()

	// Block forever
	select {}
}
