package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sjsafranek/find5/lib/api"
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
				response, _ := findapi.Do(&request)
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
