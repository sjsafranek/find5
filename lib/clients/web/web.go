package web

import (
	"os"
	"runtime"

	"github.com/sjsafranek/find5/lib/api"
	"github.com/sjsafranek/lemur"
	"github.com/sjsafranek/ligneous"
)

var logger = ligneous.AddLogger("server", "debug", "./log/find5")

type Client struct {
	api *api.Api
}

func (self *Client) Run(HTTP_PORT int) {
	logger.Debug("GOOS: ", runtime.GOOS)
	logger.Debug("CPUS: ", runtime.NumCPU())
	logger.Debug("PID: ", os.Getpid())
	logger.Debug("Go Version: ", runtime.Version())
	logger.Debug("Go Arch: ", runtime.GOARCH)
	logger.Debug("Go Compiler: ", runtime.Compiler)
	logger.Debug("NumGoroutine: ", runtime.NumGoroutine())

	server, _ := lemur.NewServer(ligneous.AddLogger("server", "debug", "./log/find5"))
	server.AttachFileServer("/static/", "static")

	authMiddleware := NewAuthenticationHandlers("chocolate-chip", self.api)

	// Sessions
	server.AttachHandlerFunc(lemur.ApiRoute{
		Name:        "login",
		Methods:     []string{"GET", "POST"},
		Pattern:     "/login",
		HandlerFunc: authMiddleware.LoginHandler("/profile"),
	})

	server.AttachHandlerFunc(lemur.ApiRoute{
		Name:        "logout",
		Methods:     []string{"GET"},
		Pattern:     "/logout",
		HandlerFunc: authMiddleware.LogoutHandler("/login"),
	})

	server.Router.Use(authMiddleware.SessionMiddleware("/login", []string{"^/$", "^/ping", "^/login", "^/logout", "^/static/*", "^/api/v1/find"}))
	//.end

	server.AttachHandlerFunc(lemur.ApiRoute{
		Name:        "index",
		Methods:     []string{"GET"},
		Pattern:     "/",
		HandlerFunc: indexHandler,
	})

	// TODO make this default???
	server.AttachHandlerFunc(lemur.ApiRoute{
		Name:        "ping",
		Methods:     []string{"GET"},
		Pattern:     "/ping",
		HandlerFunc: pingHandler,
	})

	server.AttachHandlerFunc(lemur.ApiRoute{
		Name:        "ping",
		Methods:     []string{"GET"},
		Pattern:     "/profile",
		HandlerFunc: newProfileHandler(authMiddleware),
	})

	// Api routes
	server.AttachHandlerFunc(lemur.ApiRoute{
		Name:        "api",
		Methods:     []string{"POST"},
		Pattern:     "/api/v1/find",
		HandlerFunc: newApiHandler(self.api),
	})

	server.ListenAndServe(HTTP_PORT)

}

func New(findapi *api.Api) *Client {
	return &Client{api: findapi}
}
