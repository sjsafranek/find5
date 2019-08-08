package web

import (
	"fmt"
	"net/http"
)

func apiFormatError(err error) string {
	return fmt.Sprintf(`{"status":"error","error":"%v"}`, err.Error())
}

func apiErrorResponse(w http.ResponseWriter, statusCode int, err error) {
	if 500 > statusCode {
		logger.Warn(err)
	} else {
		logger.Error(err)
	}
	apiJSONResponse(w, []byte(apiFormatError(err)), statusCode)
}

func apiBasicResponse(w http.ResponseWriter, statusCode int, err error) {
	if nil != err {
		apiErrorResponse(w, statusCode, err)
		return
	}
	apiOKResponse(w)
}

func apiOKResponse(w http.ResponseWriter) {
	apiJSONResponse(w, []byte(`{"status":"ok"}`), http.StatusOK)
}

func apiJSONResponse(w http.ResponseWriter, data []byte, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(data)
}
