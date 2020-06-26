package eventsource

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sjsafranek/logger"
	"github.com/sjsafranek/find5/findapi/lib/api"
)

// type client interface {
// 	RegisterEventListener(string, func(string,string,float64))
// }

func NewHandler(findapi *api.Api) func(http.ResponseWriter, *http.Request) {
	brokers := make(map[string]*Broker)
	return func(w http.ResponseWriter, r *http.Request) {

		// get user from authMiddleware

		vars := mux.Vars(r)
		username := vars["username"]
		if _, ok := brokers[username]; !ok {

			// create broker
			broker := NewBroker()
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
