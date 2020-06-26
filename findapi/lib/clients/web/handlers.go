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

// welcomeHandler shows a welcome message and login button.
func (self *App) indexHandler(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}

	if sessionManager.IsAuthenticated(req) {
		http.Redirect(w, req, "/profile", http.StatusFound)
		return
	}

	options := make(map[string]interface{})
	options["facebook"] = self.config.OAuth2.HasFacebook()
	options["google"] = self.config.OAuth2.HasGoogle()

	err := LOGIN_TEMPLATE.ExecuteTemplate(w, "login", options)
	if nil != err {
		logger.Error(err)
		apiBasicResponse(w, http.StatusInternalServerError, err)
	}
}

func (self *App) loginHandler(w http.ResponseWriter, r *http.Request) {

	if "POST" == r.Method {

		err := r.ParseForm()
		if nil != err {
			logger.Error(err)
			http.Error(w, "Unable to parse form", http.StatusInternalServerError)
			return
		}

		username, password, ok := r.BasicAuth()
		if !ok {
			err = errors.New("Unable to get credentials")
			apiBasicResponse(w, http.StatusBadRequest, err)
			return
		}

		results, err := self.api.Do(&api.Request{Method: "get_user", Params: &api.RequestParams{Username: username}})
		if nil != err {
			apiBasicResponse(w, http.StatusBadRequest, err)
			return
		}

		is_password, _ := results.Data.User.IsPassword(password)
		if !is_password {
			err = errors.New("Incorrect password")
			apiBasicResponse(w, http.StatusBadRequest, err)
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

	http.Redirect(w, r, "/", http.StatusFound)
}

// profileHandler shows protected user content.
func (self *App) profileHandler(w http.ResponseWriter, req *http.Request) {
	val, _ := sessionManager.Get(req)
	username := val.Values["username"].(string)
	usertype := val.Values["usertype"].(string)
	userid := val.Values["userid"].(string)
	useremail := val.Values["useremail"].(string)

	options := make(map[string]interface{})
	options["username"] = username

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
