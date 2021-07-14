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

// RefreshToken refreshes an access token
func refreshToken(ctx context.Context, logger *zap.SugaredLogger, client *Client, clientID, clientSecret, redirectURI string, tokenCh chan string, errorCh chan error, wg *sync.WaitGroup, refresh Refresh) {
	wg.Add(1)
	defer wg.Done()

	var offset = 60

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// time.Sleep(time.Second * (time.Duration(refresh.ExpiresIn) - offset))
			time.Sleep(time.Second * 10)

			logger.Debugw("refreshing token")
			t, err := client.RefreshToken(refresh, clientID, clientSecret, redirectURI)
			if err != nil {
				refresh.ExpiresIn = offset
				// TODO: what do we do if we get an unrecoverable error when refreshing?
				// write to an error channel and stop the app?
				continue
			}

			if len(t.RefreshToken) != 0 {
				logger.Debugw("got new refresh token", "refresh token", t.RefreshToken)
				refresh.RefreshToken = t.RefreshToken
			}
			if t.ExpiresIn != 0 {
				refresh.ExpiresIn = t.ExpiresIn
			}
			if len(t.AccessToken) != 0 {
				tokenCh <- t.AccessToken
			}
		}
	}
}
