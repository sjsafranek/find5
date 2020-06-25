package web

import (
	"fmt"
	"net/http"

	"github.com/sjsafranek/logger"
	"github.com/sjsafranek/lemur/middleware"
	"github.com/sjsafranek/find5/findapi/lib/api"
	"github.com/sjsafranek/find5/findapi/lib/config"
	"github.com/sjsafranek/find5/findapi/lib/socialsessions"
	// "github.com/sjsafranek/find5/findapi/lib/clients/eventsource"
	"github.com/sjsafranek/find5/findapi/lib/clients/websockets"
)

var sessionManager = socialsessions.New("chocolate-ship", "cookies")

type App struct {
	api *api.Api
	config *config.Config
	mux *http.ServeMux
}

func (self *App) ListenAndServe(address string) error {
	logger.Info(fmt.Sprintf("Magic happens on port %v...", address))
	return http.ListenAndServe(address, self.mux)
}


// New returns a new ServeMux with app routes.
// func New(findapi *api.Api, conf *config.Config) *http.ServeMux {
func New(findapi *api.Api, conf *config.Config) *App {

	app := App{
		api:findapi,
		config: conf,
		mux: http.NewServeMux(),
	}

	apiHandler := NewApiHandler(findapi, conf)

	app.mux.Handle("/", middleware.Attach(http.HandlerFunc(welcomeHandler)))
	app.mux.Handle("/profile", middleware.Attach(sessionManager.RequireLogin(http.HandlerFunc(app.profileHandler))))
	app.mux.Handle("/api", middleware.Attach(sessionManager.RequireLogin(http.HandlerFunc(apiHandler))))
	// app.mux.Handle("/api/v1", middleware.Attach(sessionManager.RequireLogin(http.HandlerFunc(apiHandler))))
	app.mux.Handle("/logout", middleware.Attach(http.HandlerFunc(sessionManager.LogoutHandler)))

	// Static files
	fsvr := http.FileServer(http.Dir("static"))
	app.mux.Handle("/static/", http.StripPrefix("/static/", fsvr))

	// get facebook login handlers
	loginHandler, callbackHandler := sessionManager.GetFacebookLoginHandlers(
		conf.Facebook.ClientID,
		conf.Facebook.ClientSecret,
		"http://localhost:8080/facebook/callback")

	// attach facebook login handlers to mux
	app.mux.Handle("/facebook/login", middleware.Attach(loginHandler))
	app.mux.Handle("/facebook/callback", middleware.Attach(callbackHandler))

	// websockets
	hub, _ := websockets.New(findapi)
	app.mux.Handle("/ws", middleware.Attach(http.HandlerFunc(hub.WebSocketHandler)))

	return &app
}
