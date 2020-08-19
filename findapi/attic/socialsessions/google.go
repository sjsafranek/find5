package socialsessions

import (
	"net/http"

	"github.com/dghubble/gologin/v2"
	"github.com/dghubble/gologin/v2/google"
	"golang.org/x/oauth2"
	googleOAuth2 "golang.org/x/oauth2/google"
)

func (self *SessionManager) GetGoogleLoginHandlers(clientID, clientSecret, callbackUrl string) (http.Handler, http.Handler) {
	// 1. Register Login and Callback handlers
	oauth2Config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  callbackUrl,
		Endpoint:     googleOAuth2.Endpoint,
		Scopes:       []string{"email"},
	}

	// state param cookies require HTTPS by default; disable for localhost development
	stateConfig := gologin.DebugOnlyCookieConfig
	loginHandler := google.StateHandler(stateConfig, google.LoginHandler(oauth2Config, nil))
	callbackHandler := google.StateHandler(stateConfig, google.CallbackHandler(oauth2Config, self.issueGoogleSession(), nil))
	return loginHandler, callbackHandler
}

// issueSession issues a cookie session after successful Facebook login
func (self *SessionManager) issueGoogleSession() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		googleUser, err := google.UserFromContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 2. Implement a success handler to issue some form of session
		session := self.issueSession()
		session.Values["userid"] = googleUser.Id
		session.Values["username"] = googleUser.Name
		session.Values["useremail"] = googleUser.Email
		session.Values["usertype"] = "google"
		session.Save(w)
		http.Redirect(w, req, "/profile", http.StatusFound)
	}
	return http.HandlerFunc(fn)
}
