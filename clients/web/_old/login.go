package web

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	// "github.com/sjsafranek/checkin/models"
)

// const cookiename = "chocolate-chip"

// var store *sessions.CookieStore

// func init() {
// secret := uuid.New().String()
// store = sessions.NewCookieStore([]byte(secret))
// }

func NewAuthenticationHandlers(cookieName string) *AuthenticationHandlers {
	secret := uuid.New().String()
	return &AuthenticationHandlers{
		store:      sessions.NewCookieStore([]byte(secret)),
		cookieName: cookieName}
}

type AuthenticationHandlers struct {
	store      *sessions.CookieStore
	cookieName string
	loginPath  string
	logoutPath string
}

func (self *AuthenticationHandlers) HasSession(r *http.Request) bool {
	session, err := self.store.Get(r, self.cookieName)

	if nil != err {
		return false
	}
	if nil == session.Values["loggedin"] {
		return false
	}
	return true
}

func (self *AuthenticationHandlers) GetUserFromSession(r *http.Request) (*models.User, error) {
	session, _ := self.store.Get(r, self.cookieName)
	username := session.Values["username"].(string)
	return models.GetUserFromUsername(username)
}

func (self *AuthenticationHandlers) LoginHandler(redirectUrl string) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

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

			user, err := models.GetUserFromUsername(username)
			if nil != err {
				apiBasicResponse(w, http.StatusBadRequest, err)
				return
			}

			is_password, _ := user.IsPassword(password)
			if !is_password {
				err = errors.New("Incorrect password")
				apiBasicResponse(w, http.StatusBadRequest, err)
				return
			}

			// create session
			session, _ := self.store.Get(r, self.cookieName)
			session.Values["loggedin"] = true
			session.Values["username"] = username
			session.Save(r, w)

			apiOKResponse(w)
			return
		}

		if self.HasSession(r) {
			http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
	<head>

		<meta charset="utf-8">
	    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
	    <meta name="description" content="">
	    <meta name="author" content="">
	    <link rel="icon" href="/docs/4.0/assets/img/favicons/favicon.ico">

		<script src="/static/jquery/3.4.1/jquery.min.js"></script>

		<!-- Font Awesome CSS -->
		<link rel="stylesheet" href="/static/fontawesome/5.9.0/css/all.css">

		<!-- Bootstrap -->
        <link rel="stylesheet" href="/static/bootstrap/4.1.3/css/bootstrap.min.css">
        <script src="/static/bootstrap/4.1.3/js/bootstrap.min.js"></script>

		<style>
			html,
			body {
				height: 100%;
			}

			body {
				display: -ms-flexbox;
				display: -webkit-box;
				display: flex;
				-ms-flex-align: center;
				-ms-flex-pack: center;
				-webkit-box-align: center;
				align-items: center;
				-webkit-box-pack: center;
				justify-content: center;
				padding-top: 40px;
				padding-bottom: 40px;
				background-color: #f5f5f5;
			}

			.form-signin {
				width: 100%;
				max-width: 330px;
				padding: 15px;
				margin: 0 auto;
			}
			.form-signin .checkbox {
				font-weight: 400;
			}
			.form-signin .form-control {
				position: relative;
				box-sizing: border-box;
				height: auto;
				padding: 10px;
				font-size: 16px;
			}
			.form-signin .form-control:focus {
				z-index: 2;
			}
			.form-signin input[type="email"] {
				margin-bottom: -1px;
				border-bottom-right-radius: 0;
				border-bottom-left-radius: 0;
			}
			.form-signin input[type="password"] {
				margin-bottom: 10px;
				border-top-left-radius: 0;
				border-top-right-radius: 0;
			}
		</style>

	</head>
	<body class="text-center">

		<div class="form-signin">

			<h1 class="h3 mb-3 font-weight-normal">Login</h1>

			<label for="username" class="sr-only">Username</label>
			<input class="form-control form-control-sm" id="username" name="username" placeholder="username" type="text" required autofocus />

			<label for="password" class="sr-only">Password</label>
			<input class="form-control form-control-sm" id="password" name="password" placeholder="password" type="password" required />

			<div class="checkbox mb-3">
	           <label>
	             <input type="checkbox" value="remember-me"> Remember me
	           </label>
	         </div>

			<button id="login" class="btn btn-sm btn-primary btn-block" >
				<i class="fas fa-sign-in-alt"></i>
				Login
			</button>

			<p class="mt-5 mb-3 text-muted">&copy; 2019-2020</p>

			<div class='message'></div>
		</div>

		<script>
			function login(){
				var usr = $('#username').val();
				var psw = $('#password').val();
				$.ajax({
					method: 'POST',
					url: '/login',
					username: usr,
					password: psw,
					headers: {
						Authorization: 'Basic ' + btoa(usr+':'+psw)
					}
				}).done(function(data){
					console.log(data);
					window.location = '`+redirectUrl+`'
				}).fail(function(xhr){
					console.log(xhr.responseText);
					$('.message').append(xhr.responseText);
				});
			}
			$("#login").on('click', login);
		</script>
	</body>
</html>`)
	}
}

func (self *AuthenticationHandlers) LogoutHandler(redirectUrl string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// delete session
		session, _ := self.store.Get(r, self.cookieName)
		session.Options.MaxAge = -1
		session.Save(r, w)

		http.Redirect(w, r, redirectUrl, http.StatusTemporaryRedirect)
	}
}

func (self *AuthenticationHandlers) SessionMiddleware(redirectUrl string, unprotected []string) func(http.Handler) http.Handler {
	// return middleware handler
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//
			matched := false
			for _, pattern := range unprotected {
				if !matched {
					matched, _ = regexp.MatchString(pattern, r.URL.Path)
				}
			}

			// redirected if url is protected
			if !self.HasSession(r) && !matched {
				http.Redirect(w, r, redirectUrl, http.StatusTemporaryRedirect)
				return
			}

			// go to next handler
			next.ServeHTTP(w, r)
		})
	}
}
