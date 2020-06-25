package web

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/sjsafranek/logger"
	"github.com/sjsafranek/find5/findapi/lib/api"
	"github.com/sjsafranek/find5/findapi/lib/config"
)

func NewApiHandler(rpcApi *api.Api, conf *config.Config) func(w http.ResponseWriter, r *http.Request) {

// func apiHandler(w http.ResponseWriter, r *http.Request) {
return func(w http.ResponseWriter, r *http.Request) {
	var data string

	val, _ := sessionManager.Get(r)
	useremail := val.Values["useremail"].(string)
	if 0 == len(useremail) {
		apiBasicResponse(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}

	status_code, err := func() (int, error) {
		switch r.Method {
		case "POST":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return http.StatusBadRequest, err
			}
			r.Body.Close()

			var request api.Request

			// TODO
			//  - use request.Unmarshal
			json.Unmarshal(body, &request)

			// WARNING - protect against request hijacking!
			if nil == request.Params {
				request.Params = &api.RequestParams{}
			}
			request.Params.Username = useremail
			request.Params.Apikey = ""
			//.END

			logger.Info(string(body))
			// logger.Info(request)

			if !conf.Api.IsPublicMethod(request.Method) {
				logger.Warnf("Not a public api method: %v", request.Method)
				return http.StatusMethodNotAllowed, errors.New(http.StatusText(http.StatusMethodNotAllowed))
			}

			// run api request
			response, err := rpcApi.Do(&request)

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
