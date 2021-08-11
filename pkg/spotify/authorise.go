package spotify

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// func authorise() - use this to make request for auth? need to print out the URL for the user to auth with and trigger callback

// GetToken gets a token using supplied auth code and client data
func GetToken(ctx context.Context, logger *zap.SugaredLogger, authCode, clientID, clientSecret, redirectURI string, tokenCh chan string, errorCh chan error, wg *sync.WaitGroup) error {
	var t Token
	var r Refresh

	logger.Debugw("getting token")
	client := NewClient("https://accounts.spotify.com/api/token")
	t, err := client.GetToken(authCode, clientID, clientSecret, redirectURI)
	if err != nil {
		return fmt.Errorf("error getting token: %v", err)
	}

	tokenCh <- t.AccessToken

	r.RefreshToken = t.RefreshToken
	r.ExpiresIn = t.ExpiresIn
	go refreshToken(ctx, logger, client, clientID, clientSecret, redirectURI, tokenCh, errorCh, wg, r)

	return nil
}

// refreshToken refreshes an access token
func refreshToken(ctx context.Context, logger *zap.SugaredLogger, client *Client, clientID, clientSecret, redirectURI string, tokenCh chan string, errorCh chan error, wg *sync.WaitGroup, r Refresh) {
	logger.Debugw("refresh routine started")

	wg.Add(1)
	defer wg.Done()

	retryCh := make(chan int, 1)
	var offset = 60

	ticker := time.NewTicker(time.Duration(r.ExpiresIn-offset) * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			logger.Debugw("ticker refreshing token")
			refresh(logger, client, &r, clientID, clientSecret, redirectURI, tokenCh, retryCh)
		case <-retryCh:
			logger.Debugw("retry channel refreshing token")
			refresh(logger, client, &r, clientID, clientSecret, redirectURI, tokenCh, retryCh)
		default:
			continue
		}
	}
}

func refresh(logger *zap.SugaredLogger, client *Client, r *Refresh, clientID, clientSecret, redirectURI string, tokenCh chan string, retryCh chan int) {
	logger.Debugw("refreshing token")
	t, err := client.RefreshToken(*r, clientID, clientSecret, redirectURI)
	if err != nil {
		retryCh <- 1
	}

	if len(t.RefreshToken) != 0 {
		logger.Debugw("got new refresh token", "refresh token", t.RefreshToken)
		r.RefreshToken = t.RefreshToken
	}
	if len(t.AccessToken) != 0 {
		tokenCh <- t.AccessToken
	}

}
