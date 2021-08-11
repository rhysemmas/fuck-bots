package playlist

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/rhysemmas/fuck-bots/pkg/http"
	"github.com/rhysemmas/fuck-bots/pkg/spotify"
)

type protector struct {
	ctx          context.Context
	clientID     string
	clientSecret string
	playlistID   string
	playlistName string
	redirectURI  string
	tokenCh      chan string
	errorCh      chan error
	waitGroup    *sync.WaitGroup
	logger       *zap.SugaredLogger
}

// NewProtector protects playlists
func NewProtector(logger *zap.SugaredLogger, addr, clientID, clientSecret, playlistID, playlistName, redirectURI string) error {
	var wg sync.WaitGroup
	tokenCh := make(chan string, 1)
	errorCh := make(chan error, 1)
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	p := protector{
		ctx:          ctx,
		clientID:     clientID,
		clientSecret: clientSecret,
		playlistID:   playlistID,
		playlistName: playlistName,
		redirectURI:  redirectURI,
		tokenCh:      tokenCh,
		errorCh:      errorCh,
		waitGroup:    &wg,
		logger:       logger,
	}

	go p.protectPlaylist()
	go p.startCallbackServer(addr)

	select {
	case signal := <-stopCh:
		logger.Infow("shutdown signal received", "signal", signal)
		cancel()
	case err := <-errorCh:
		logger.Warnw("fatal error, stopping", "err", err)
		cancel()
	}

	wg.Wait()

	return fmt.Errorf("all go routines have exited")
}

func (p *protector) startCallbackServer(addr string) {
	p.waitGroup.Add(1)
	defer p.waitGroup.Done()

	routes := http.NewRoutes(p.ctx, p.logger, p.clientID, p.clientSecret, p.redirectURI, p.tokenCh, p.errorCh, p.waitGroup)
	httpShutdown := http.NewServer(addr, routes, p.logger).Start(p.errorCh)
	defer httpShutdown(p.ctx)

	p.logger.Infow("ready to serve", "address", addr)
	select {
	case <-p.ctx.Done():
		return
	}
}

func (p *protector) protectPlaylist() {
	p.waitGroup.Add(1)
	defer p.waitGroup.Done()

	var token string
	client := spotify.NewClient("https://api.spotify.com/v1/playlists/" + p.playlistID)

	for {
		select {
		case token = <-p.tokenCh:
			p.logger.Debugw("got new token", "token", token)
		case <-p.ctx.Done():
			return
		default:
			if len(token) == 0 {
				continue
			}

			p.logger.Debugw("making playlist request")

			playlist, err := client.GetPlaylistDetails(token)
			if err != nil {
				p.logger.Warnw("error getting playlist details", "error", err)
			}

			p.logger.Infow("got playlist", "playlist", playlist)

			if playlist.Name != p.playlistName {
				var updatedPlaylist spotify.Playlist
				updatedPlaylist.Name = p.playlistName

				p.logger.Infow("updating playlist", "playlist", updatedPlaylist)
				err := client.UpdatePlaylistDetails(token, updatedPlaylist)
				if err != nil {
					p.logger.Warnw("error updating playlist", "error", err)
				}
			}

			time.Sleep(5 * time.Second)
		}
	}
}
