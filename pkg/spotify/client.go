package spotify

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	backoffv4 "github.com/cenkalti/backoff/v4"
)

// RateLimitError is an error used when spotify rate limits requests
type RateLimitError struct {
	E string
}

// Error returns the error string for RateLimitError
func (r *RateLimitError) Error() string {
	return r.E
}

// Client contains the endpoint that API calls will be made to
type Client struct {
	Endpoint string
}

// NewClient creates a new spotify client
func NewClient(endpoint string) *Client {
	return &Client{
		Endpoint: endpoint,
	}
}

func (c *Client) Authorise(clientID, redirectURI string) (string, error) {
	var location string

	err := RetryAPICall(func() (waitTime int, err error) {
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		now := time.Now().Format(time.RFC3339)
		h := sha1.New()
		h.Write([]byte(now))
		sha := h.Sum(nil)

		params := "client_id=" + url.QueryEscape(clientID) +
			"&response_type=code" +
			"&redirect_uri=" + url.QueryEscape(redirectURI) +
			"&scope=playlist-modify-public&playlist-modify-private" +
			"&state=" + string(sha)

		path := fmt.Sprintf(c.Endpoint+"?%s", params)

		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			return 0, fmt.Errorf("error creating http request: %v", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return 0, fmt.Errorf("error making authorisation request: %v", err)
		}

		defer resp.Body.Close()

		if resp.StatusCode == 429 {
			waitTime, err := strconv.Atoi(resp.Header["Retry-After"][0])
			if err != nil {
				return 0, fmt.Errorf("error converting retry after to int: %v", err)
			}
			return waitTime, &RateLimitError{E: "hit rate limit"}
		} else if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			return 0, fmt.Errorf("got http error code: %v", resp.StatusCode)
		}

		location = resp.Header.Get("location")
		if location == "" {
			return 0, fmt.Errorf("got empty location header")
		}

		return 0, nil
	})

	return location, err
}

// GetToken gets a token from the spotify API
func (c *Client) GetToken(authCode, clientID, clientSecret, redirectURI string) (Token, error) {
	var t Token

	err := RetryAPICall(func() (waitTime int, err error) {
		resp, err := http.PostForm(c.Endpoint,
			url.Values{
				"client_id":     {clientID},
				"client_secret": {clientSecret},
				"grant_type":    {"authorization_code"},
				"code":          {authCode},
				"redirect_uri":  {redirectURI},
			})
		if err != nil {
			return 0, fmt.Errorf("error making token request: %v", err)
		}

		defer resp.Body.Close()

		if resp.StatusCode == 429 {
			waitTime, err := strconv.Atoi(resp.Header["Retry-After"][0])
			if err != nil {
				return 0, fmt.Errorf("error converting retry after to int: %v", err)
			}
			return waitTime, &RateLimitError{E: "hit rate limit"}
		} else if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			return 0, fmt.Errorf("got http error code: %v", resp.StatusCode)
		}

		if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
			return 0, fmt.Errorf("error occured trying to read request body: %v", err)
		}

		return 0, nil
	})

	return t, err
}

// RefreshToken refreshes a token for use with the spotify API
func (c *Client) RefreshToken(refresh Refresh, clientID, clientSecret, redirectURI string) (Token, error) {
	var t Token

	err := RetryAPICall(func() (waitTime int, err error) {
		resp, err := http.PostForm(c.Endpoint,
			url.Values{
				"client_id":     {clientID},
				"client_secret": {clientSecret},
				"grant_type":    {"refresh_token"},
				"refresh_token": {refresh.RefreshToken},
			})
		if err != nil {
			return 0, fmt.Errorf("error making refresh request: %v", err)
		}

		defer resp.Body.Close()

		if resp.StatusCode == 429 {
			waitTime, err := strconv.Atoi(resp.Header["Retry-After"][0])
			if err != nil {
				return 0, fmt.Errorf("error converting retry after to int: %v", err)
			}
			return waitTime, &RateLimitError{E: "hit rate limit"}
		} else if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			return 0, fmt.Errorf("got http error code: %v", resp.StatusCode)
		}

		if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
			return 0, fmt.Errorf("error occured trying to read request body: %v", err)
		}

		return 0, nil
	})

	return t, err
}

// GetPlaylistDetails gets the current state of a playlist
func (c *Client) GetPlaylistDetails(token string) (Playlist, error) {
	var p Playlist

	err := RetryAPICall(func() (waitTime int, err error) {
		client := &http.Client{}

		req, err := http.NewRequest("GET", c.Endpoint, nil)
		if err != nil {
			return 0, fmt.Errorf("error creating http request: %v", err)
		}

		req.Header.Add("Authorization", "Bearer "+token)
		resp, err := client.Do(req)
		if err != nil {
			return 0, fmt.Errorf("error making refresh request: %v", err)
		}

		defer resp.Body.Close()

		if resp.StatusCode == 429 {
			waitTime, err := strconv.Atoi(resp.Header["Retry-After"][0])
			if err != nil {
				return 0, fmt.Errorf("error converting retry after to int: %v", err)
			}
			return waitTime, &RateLimitError{E: "hit rate limit"}
		} else if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			return 0, fmt.Errorf("got http error code: %v", resp.StatusCode)
		}

		if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
			return 0, fmt.Errorf("error occured trying to read request body: %v", err)
		}

		return 0, nil
	})

	return p, err
}

// UpdatePlaylistDetails updates the state of a playlist
func (c *Client) UpdatePlaylistDetails(token string, playlist Playlist) error {
	err := RetryAPICall(func() (waitTime int, err error) {
		client := &http.Client{}

		requestBody, err := json.Marshal(playlist)
		if err != nil {
			return 0, fmt.Errorf("error marshalling details to json: %v", err)
		}

		req, err := http.NewRequest("PUT", c.Endpoint, bytes.NewBuffer(requestBody))
		if err != nil {
			return 0, fmt.Errorf("error creating http request: %v", err)
		}

		req.Header.Add("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		resp, err := client.Do(req)
		if err != nil {
			return 0, fmt.Errorf("error making refresh request: %v", err)
		}

		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0, fmt.Errorf("error reading response body: %v", err)
		}

		if resp.StatusCode == 429 {
			waitTime, err := strconv.Atoi(resp.Header["Retry-After"][0])
			if err != nil {
				return 0, fmt.Errorf("error converting retry after to int: %v", err)
			}
			return waitTime, &RateLimitError{E: "hit rate limit"}
		} else if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			return 0, fmt.Errorf("got http error code: %v, body: %v", resp.StatusCode, b)
		}

		return 0, nil
	})

	return err
}

// RetryAPICall retries calls to spotify API
func RetryAPICall(operation func() (int, error)) error {
	var waitTime int
	op := func() (err error) {
		waitTime, err = operation()
		if _, ok := err.(*RateLimitError); ok {
			return err
		}

		if err != nil {
			return backoffv4.Permanent(err)
		}

		return nil
	}

	return backoffv4.Retry(op, backoffv4.NewConstantBackOff(time.Second*time.Duration(waitTime)))
}
