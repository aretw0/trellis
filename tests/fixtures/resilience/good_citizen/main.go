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
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	fmt.Fprintln(os.Stderr, "Good Citizen Started")

	done := make(chan bool, 1)

	go func() {
		sig := <-sigs
		fmt.Fprintf(os.Stderr, "\nReceived signal: %s\n", sig)
		fmt.Fprintln(os.Stderr, "Cleaning up...")
		time.Sleep(500 * time.Millisecond) // Simulate cleanup work
		fmt.Fprintln(os.Stderr, "Cleanup done")
		done <- true
	}()

	// Simulate long running work
	select {
	case <-done:
		fmt.Println("Exiting gracefully")
		os.Exit(0)
	case <-time.After(10 * time.Second):
		fmt.Println("Finished work (timeout)")
	}
}
