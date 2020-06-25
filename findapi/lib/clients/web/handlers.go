package web

import (
	"github.com/sjsafranek/logger"
	"html/template"
	"net/http"
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
