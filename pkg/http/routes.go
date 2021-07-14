package http

import (
	"context"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/rhysemmas/fuck-bots/pkg/spotify"
)

// Routes is a struct containing an http router and logger
type Routes struct {
	*mux.Router
	ctx          context.Context
	logger       *zap.SugaredLogger
	clientID     string
	clientSecret string
	redirectURI  string
	tokenCh      chan string
	errorCh      chan error
	waitGroup    *sync.WaitGroup
}

// NewRoutes returns an http handler with routes configured
func NewRoutes(ctx context.Context, logger *zap.SugaredLogger, clientID, clientSecret, redirectURI string, tokenCh chan string, errorCh chan error, wg *sync.WaitGroup) http.Handler {
	router := mux.NewRouter()
	routes := Routes{
		Router:       router,
		ctx:          ctx,
		logger:       logger,
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		tokenCh:      tokenCh,
		errorCh:      errorCh,
		waitGroup:    wg,
	}
	routes.setup()

	return routes
}

func (r *Routes) setup() {
	r.HandleFunc("/status", r.readinessHandler).Methods("GET")
	r.HandleFunc("/spotify/callback", r.callbackHandler).Methods("GET", "POST")
}

func (r *Routes) readinessHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// authHandler receives and processes a spotify OAuth code
func (r *Routes) callbackHandler(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		r.logger.Warnw("error occured trying to parse form", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	codeSlice := req.Form["code"]
	authCode := codeSlice[0]
	r.logger.Debugw("got auth code", "code", authCode)

	err := spotify.GetToken(r.ctx, r.logger, authCode, r.clientID, r.clientSecret, r.redirectURI, r.tokenCh, r.errorCh, r.waitGroup)
	if err != nil {
		r.logger.Warnw("error occured trying to get token", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
