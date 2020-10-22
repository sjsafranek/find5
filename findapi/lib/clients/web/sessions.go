package web

import (
	"fmt"
	"net/http"

	"github.com/sjsafranek/gosocialsessions"
	"github.com/sjsafranek/logger"
	"github.com/sjsafranek/lemur/middleware"

	"github.com/sjsafranek/find5/findapi/lib/api"
	"github.com/sjsafranek/find5/findapi/lib/config"
	// "github.com/sjsafranek/find5/findapi/lib/clients/eventsource"
	"github.com/sjsafranek/find5/findapi/lib/clients/websockets"
)

var sessionManager = gosocialsessions.New("chocolate-ship", "cookies")

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
func New(findapi *api.Api, conf *config.Config) *App {

	app := App{
		api:findapi,
		config: conf,
		mux: http.NewServeMux(),
	}

	app.mux.Handle("/", middleware.Attach(http.HandlerFunc(app.indexHandler)))
	app.mux.Handle("/login", middleware.Attach(http.HandlerFunc(app.indexHandler)))
	app.mux.Handle("/logout", middleware.Attach(http.HandlerFunc(sessionManager.LogoutHandler)))
	app.mux.Handle("/profile", middleware.Attach(sessionManager.RequireLogin(http.HandlerFunc(app.profileHandler))))
	app.mux.Handle("/api", middleware.Attach(sessionManager.RequireLogin(http.HandlerFunc(app.apiWithSessionHandler))))
	app.mux.Handle("/api/v1", middleware.Attach(http.HandlerFunc(app.apiHandler)))

	// Static files
	fsvr := http.FileServer(http.Dir("static"))
	app.mux.Handle("/static/", http.StripPrefix("/static/", fsvr))

	// Enable FaceBook login
	if conf.OAuth2.HasFacebook() {
		// get facebook login handlers
		loginHandler, callbackHandler := sessionManager.GetFacebookLoginHandlers(
			conf.OAuth2.Facebook.ClientID,
			conf.OAuth2.Facebook.ClientSecret,
			// "http://localhost:8080/facebook/callback")
			fmt.Sprintf("%v/facebook/callback", conf.Server.GetURLString()))

		// attach facebook login handlers to mux
		app.mux.Handle("/facebook/login", middleware.Attach(loginHandler))
		app.mux.Handle("/facebook/callback", middleware.Attach(callbackHandler))
	}

	if conf.OAuth2.HasGoogle() {
		// get facebook login handlers
		loginHandler, callbackHandler := sessionManager.GetGoogleLoginHandlers(
			conf.OAuth2.Google.ClientID,
			conf.OAuth2.Google.ClientSecret,
			// "http://localhost:8080/google/callback")
			fmt.Sprintf("%v/google/callback", conf.Server.GetURLString()))

		// attach facebook login handlers to mux
		app.mux.Handle("/google/login", middleware.Attach(loginHandler))
		app.mux.Handle("/google/callback", middleware.Attach(callbackHandler))
	}

	// websockets
	hub, _ := websockets.New(findapi)
	app.mux.Handle("/ws", middleware.Attach(http.HandlerFunc(hub.WebSocketHandler)))

	return &app
}
