package web

import (
	"errors"
	"html/template"
	"net/http"

	"github.com/sjsafranek/logger"
	"github.com/sjsafranek/find5/findapi/lib/api"
)

var (
	LOGIN_TEMPLATE   *template.Template = template.Must(template.ParseFiles("tmpl/global_header.html", "tmpl/global_footer.html", "tmpl/login.html"))

	PROFILE_TEMPLATE *template.Template = template.Must(template.ParseFiles("tmpl/global_header.html", "tmpl/global_footer.html", "tmpl/navbar.html", "tmpl/profile.html"))
)

func (self *App) executeLoginTemplate(w http.ResponseWriter, options map[string]interface{}) {
	logger.Info(options)
	err := LOGIN_TEMPLATE.ExecuteTemplate(w, "login", options)
	if nil != err {
		logger.Error(err)
		apiBasicResponse(w, http.StatusInternalServerError, err)
	}
}

func (self *App) getHandlerOptions(r *http.Request) map[string]interface{} {
	options := make(map[string]interface{})

	oauth2Options := make(map[string]bool)
	oauth2Options["facebook"] = self.config.OAuth2.HasFacebook()
	oauth2Options["google"] = self.config.OAuth2.HasGoogle()
	options["oauth2"] = oauth2Options

	val, _ := sessionManager.Get(r)
	if nil != val {
		useremail := val.Values["useremail"]
		username := useremail.(string)
		userOptions := make(map[string]string)
		userOptions["username"] = username
		results, err := self.api.Do(&api.Request{Method: "get_user", Params: &api.RequestParams{Username: username}})
		if nil == err {
			userOptions["apikey"] = results.Data.User.Apikey
		}

		options["user"] = userOptions
	}

	return options
}

// welcomeHandler shows a welcome message and login button.
func (self *App) indexHandler(w http.ResponseWriter, r *http.Request) {

	if sessionManager.IsAuthenticated(r) {
		http.Redirect(w, r, "/profile", http.StatusFound)
		return
	}

	options := self.getHandlerOptions(r)

	if "POST" == r.Method {

		username := r.FormValue("username")
		password := r.FormValue("password")
		if "" == username && "" == password {
			usr, psw, ok := r.BasicAuth()
			if !ok {
				err := errors.New("Unable to get credentials")
				options["error"] = err.Error()
				self.executeLoginTemplate(w, options)
				return
			}
			username = usr
			password = psw
		}

		results, err := self.api.Do(&api.Request{Method: "get_user", Params: &api.RequestParams{Username: username}})
		if nil != err {
			options["error"] = err.Error()
			self.executeLoginTemplate(w, options)
			return
		}

		is_password, _ := results.Data.User.IsPassword(password)
		if !is_password {
			err = errors.New("Incorrect password")
			options["error"] = err.Error()
			self.executeLoginTemplate(w, options)
			return
		}

		session := sessionManager.IssueSession()
		session.Values["userid"] = ""
		session.Values["username"] = results.Data.User.Username
		session.Values["useremail"] = results.Data.User.Email
		session.Values["usertype"] = "find5"
		session.Save(w)
		http.Redirect(w, r, "/profile", http.StatusFound)
		return
	}

	self.executeLoginTemplate(w, options)
}

// profileHandler shows protected user content.
func (self *App) profileHandler(w http.ResponseWriter, r *http.Request) {
	val, _ := sessionManager.Get(r)
	username := val.Values["username"].(string)
	usertype := val.Values["usertype"].(string)
	userid := val.Values["userid"].(string)
	useremail := val.Values["useremail"].(string)

	// options := make(map[string]interface{})
	// options["username"] = username
	options := self.getHandlerOptions(r)

	user, err := self.api.GetDatabase().CreateUserIfNotExists(useremail, useremail)
	if nil != err {
		logger.Error(err)
		apiBasicResponse(w, http.StatusInternalServerError, err)
	}
	user.CreateSocialAccountIfNotExists(userid, username, usertype)

	err = PROFILE_TEMPLATE.ExecuteTemplate(w, "profile", options)
	if nil != err {
		logger.Error(err)
		apiBasicResponse(w, http.StatusInternalServerError, err)
	}
}
