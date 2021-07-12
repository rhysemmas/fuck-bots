package main

import (
	"flag"
	"fmt"
	"os"

	"go.uber.org/zap"
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
	flag.IntVar(&rpm, "key", defaultRpm, "Rate of requests to Spotify API per minute, defaults to 60")
	flag.Parse()

	if err := exec(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func exec() error {
	logger, err := initialiseLogger()
	if err != nil {
		return fmt.Errorf("unable to initialise logger, %w", err)
	}
	defer logger.Sync()

	return nil
}

func initialiseLogger() (*zap.SugaredLogger, error) {
	l, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	return l.Sugar(), nil
}
