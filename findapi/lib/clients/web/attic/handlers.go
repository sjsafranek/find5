package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sjsafranek/find5/findapi/lib/api"
)

var startTime time.Time

func init() {
	startTime = time.Now()
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	apiOKResponse(w)
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	response := fmt.Sprintf(`{"status":"ok","data":{"message":"pong","runtime":%v,"registered":"%v"}}`, time.Since(startTime).Seconds(), startTime)
	apiJSONResponse(w, []byte(response), http.StatusOK)
}

func newApiHandler(findapi *api.Api) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var data string
		status_code, err := func() (int, error) {
			switch r.Method {
			case "POST":
				body, err := ioutil.ReadAll(r.Body)
				if err != nil {
					return http.StatusBadRequest, err
				}
				r.Body.Close()

				var request api.Request
				json.Unmarshal(body, &request)

				// TODO
				// // protect against hijacking!
				// request.Params.Username = string(useremail[0])
				// request.Params.Apikey = ""

				// run api request
				response, err := findapi.Do(&request)

				if nil != err {
					results, _ := response.Marshal()
					data = results
					return http.StatusBadRequest, nil
				}

				results, _ := response.Marshal()
				data = results
				return http.StatusOK, nil
			default:
				return http.StatusMethodNotAllowed, errors.New(http.StatusText(http.StatusMethodNotAllowed))
			}
		}()

		if nil != err {
			apiBasicResponse(w, status_code, err)
			return
		}

		apiJSONResponse(w, []byte(data), status_code)
	}
}

func newProfileHandler(authMiddleware *AuthenticationHandlers) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := authMiddleware.GetUserFromSession(r)
		if nil != err {
			apiBasicResponse(w, http.StatusInternalServerError, err)
			return
		}

		t := template.Must(template.ParseFiles("tmpl/profile.html"))
		t.Delims("[[", "]]")
		err = t.ExecuteTemplate(w, "profile", data.Data.User)
		if nil != err {
			apiBasicResponse(w, http.StatusInternalServerError, err)
			return
		}
	}
}
