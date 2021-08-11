package main

import (
	"flag"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/rhysemmas/fuck-bots/pkg/playlist"
)

var (
	defaultAddr         = "0.0.0.0"
	defaultPort         = 8080
	defaultPlaylistID   = ""
	defaultPlaylistName = ""
	defaultClientID     = ""
	defaultClientSecret = ""
	defaultRedirectURI  = ""
	defaultDebug        = false

	addr         string
	port         int
	playlistID   string
	playlistName string
	clientID     string
	clientSecret string
	redirectURI  string
	debug        bool
)

func main() {
	flag.StringVar(&addr, "address", defaultAddr, "Address to run http server, defaults to 0.0.0.0")
	flag.IntVar(&port, "port", defaultPort, "Port to run http server, defaults to 8080")
	flag.StringVar(&playlistID, "playlist-id", defaultPlaylistID, "ID of playlist to protect")
	flag.StringVar(&playlistName, "playlist-name", defaultPlaylistName, "Expected name of the playlist being protected")
	flag.StringVar(&clientID, "client-id", defaultClientID, "Spotify app client ID, defaults to value of CLIENT_ID env var")
	flag.StringVar(&clientSecret, "client-secret", defaultClientSecret, "Spotify app client secret, defaults to value of CLIENT_SECRET env var")
	flag.StringVar(&redirectURI, "redirect-uri", defaultRedirectURI, "Redirect URI for completing Spotify OAuth")
	flag.BoolVar(&debug, "debug", defaultDebug, "Debug mode, defaults to false")

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

	addr = fmt.Sprintf("%s:%d", addr, port)

	if err = playlist.NewProtector(logger, addr, clientID, clientSecret, playlistID, playlistName, redirectURI); err != nil {
		return fmt.Errorf("error protecting playlist: %v", err)
	}

	return nil
}

func initialiseLogger() (*zap.SugaredLogger, error) {
	var l *zap.Logger
	var err error

	if debug {
		l, err = zap.NewDevelopment()
	} else {
		l, err = zap.NewProduction()
	}
	if err != nil {
		return nil, err
	}

	return l.Sugar(), nil
}
