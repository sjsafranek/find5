package web

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/sjsafranek/logger"
	"github.com/sjsafranek/find5/findapi/lib/api"
)


func (self *App) getApiRequestFromHttpRequest(r *http.Request) (*api.Request, error) {
	var request api.Request

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return &request, err
	}
	r.Body.Close()

	logger.Debug(string(body))

	// TODO
	//  - use request.Unmarshal
	return &request, json.Unmarshal(body, &request)
}

func (self *App) execApiRequest(request *api.Request) (string, int, error) {
	// run api request
	response, err := self.api.Do(request)
	results, _ := response.Marshal()

	if nil != err {
		return results, http.StatusBadRequest, nil
	}

	return results, http.StatusOK, nil
}

func (self *App) apiHandler(w http.ResponseWriter, r *http.Request) {

	var data string

	status_code, err := func() (int, error) {
		switch r.Method {
		case "POST":
			// get api request
			request, err := self.getApiRequestFromHttpRequest(r)
			if err != nil {
				return http.StatusBadRequest, err
			}

			// check against allowed methods
			if !self.api.IsPublicMethod(request.Method) {
				logger.Warnf("Not a public api method: %v", request.Method)
				return http.StatusMethodNotAllowed, errors.New(http.StatusText(http.StatusMethodNotAllowed))
			}

			// run api request
			results, statusCode, err := self.execApiRequest(request)
			data = results
			return statusCode, err

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


func (self *App) apiWithSessionHandler(w http.ResponseWriter, r *http.Request) {

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
			// get api request
			request, err := self.getApiRequestFromHttpRequest(r)
			if err != nil {
				return http.StatusBadRequest, err
			}

			// WARNING - protect against request hijacking!
			if nil == request.Params {
				request.Params = &api.RequestParams{}
			}
			request.Params.Username = useremail
			request.Params.Apikey = ""

			// check against allowed methods
			if !self.api.IsPublicMethod(request.Method) {
				logger.Warnf("Not a public api method: %v", request.Method)
				return http.StatusMethodNotAllowed, errors.New(http.StatusText(http.StatusMethodNotAllowed))
			}

			// run api request
			results, statusCode, err := self.execApiRequest(request)
			data = results
			return statusCode, err

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
