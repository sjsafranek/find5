package socialsessions

import (
	"net/http"

	"github.com/dghubble/gologin/v2"
	"github.com/dghubble/gologin/v2/facebook"
	// "github.com/dghubble/gologin"
	// "github.com/dghubble/gologin/facebook"
	"golang.org/x/oauth2"
	facebookOAuth2 "golang.org/x/oauth2/facebook"
)

func (self *SessionManager) GetFacebookLoginHandlers(clientID, clientSecret, callbackUrl string) (http.Handler, http.Handler) {
	// 1. Register Login and Callback handlers
	oauth2Config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  callbackUrl,
		Endpoint:     facebookOAuth2.Endpoint,
		Scopes:       []string{"email"},
	}

	// state param cookies require HTTPS by default; disable for localhost development
	stateConfig := gologin.DebugOnlyCookieConfig
	loginHandler := facebook.StateHandler(stateConfig, facebook.LoginHandler(oauth2Config, nil))
	callbackHandler := facebook.StateHandler(stateConfig, facebook.CallbackHandler(oauth2Config, self.issueFacebookSession(), nil))
	return loginHandler, callbackHandler
}

// issueSession issues a cookie session after successful Facebook login
func (self *SessionManager) issueFacebookSession() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		facebookUser, err := facebook.UserFromContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 2. Implement a success handler to issue some form of session
		session := self.issueSession()
		session.Values["userid"] = facebookUser.ID
		session.Values["username"] = facebookUser.Name
		session.Values["useremail"] = facebookUser.Email
		session.Values["usertype"] = "facebook"
		session.Save(w)
		http.Redirect(w, req, "/profile", http.StatusFound)
	}
	return http.HandlerFunc(fn)
}
