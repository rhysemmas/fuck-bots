package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	defaultKey string = os.Getenv("API_KEY")
	defaultRpm int    = 60
	key        string
	rpm        int
)

func main() {
	// get current playlist details from spotify
	// write to work channel if there is work to do
	// read from work channel, update playlist name if there is work to do
	flag.StringVar(&key, "key", defaultKey, "Spotify API key, defaults to value of API_KEY env var")
	flag.Parse()

	if err := exec(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func exec() error {

	return nil
}
