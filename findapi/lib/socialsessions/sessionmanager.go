package socialsessions

import (
	"net/http"

	"github.com/dghubble/sessions"
)

type SessionManager struct {
	sessionName   string
	sessionSecret string
	sessionStore  *sessions.CookieStore
}

func New(sessionName, sessionSecret string) *SessionManager {
	// sessionStore encodes and decodes session data stored in signed cookies
	return &SessionManager{
		sessionName:   sessionName,
		sessionSecret: sessionSecret,
		sessionStore:  sessions.NewCookieStore([]byte(sessionSecret), nil),
	}
}

func (self *SessionManager) Get(req *http.Request) (*sessions.Session, error) {
	return self.sessionStore.Get(req, self.sessionName)
}

func (self *SessionManager) issueSession() *sessions.Session {
	return self.sessionStore.New(self.sessionName)
}

func (self *SessionManager) destroySession(w http.ResponseWriter) {
	self.sessionStore.Destroy(w, self.sessionName)
}

// logoutHandler destroys the session on POSTs and redirects to home.
func (self *SessionManager) LogoutHandler(w http.ResponseWriter, req *http.Request) {
	// if req.Method == "POST" {
	self.destroySession(w)
	// }
	http.Redirect(w, req, "/", http.StatusFound)
}

// requireLogin redirects unauthenticated users to the login route.
func (self *SessionManager) RequireLogin(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		if !self.IsAuthenticated(req) {
			http.Redirect(w, req, "/", http.StatusFound)
			return
		}
		next.ServeHTTP(w, req)
	}
	return http.HandlerFunc(fn)
}

// isAuthenticated returns true if the user has a signed session cookie.
func (self *SessionManager) IsAuthenticated(req *http.Request) bool {
	if _, err := self.sessionStore.Get(req, self.sessionName); err == nil {
		return true
	}
	return false
}
