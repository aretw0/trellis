package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Crashy Started")
	fmt.Fprintln(os.Stderr, "Something went terribly wrong!")
	os.Exit(123)
}
