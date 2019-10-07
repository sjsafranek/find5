package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
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

func newEventSourceHandler(findapi *api.Api) func(http.ResponseWriter, *http.Request) {
	brokers := make(map[string]*Broker)
	return func(w http.ResponseWriter, r *http.Request) {

		// get user from authMiddleware

		vars := mux.Vars(r)
		username := vars["username"]
		if _, ok := brokers[username]; !ok {

			// create broker
			broker := NewServer()
			brokers[username] = broker
			// listen for events
			go func() {
				findapi.RegisterEventListener(username, func(device_id string, location_id string, probability float64) {
					logger.Info(device_id, location_id, probability)
					brokers[username].Notifier <- []byte(fmt.Sprintf(`{"status":"ok","data":{"device_id":"%v","location_id":"%v","probability":%v}}`, device_id, location_id, probability))
				})
			}()
			//.end
		}

		brokers[username].ServeHTTP(w, r)
	}
}
