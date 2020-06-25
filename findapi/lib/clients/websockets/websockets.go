package websockets

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sjsafranek/logger"
	"github.com/sjsafranek/find5/findapi/lib/api"
)


// type request interface {}
//
// type response interface {
// 	Marshal() (string, error)
// }
//
// type api interface {
// 	Do (request) (response, error)
// 	DoJSON (string) (response, error)
// }

var hub WebSocketHub

var upgrader = websocket.Upgrader{}

const (
	StatusServiceRestart = 1012
	StatusOK             = 4001
	StatusBadRequest     = 4002
	StatusMethodNotFound = 4003
	StatusInvalidSession = 4008
	StatusDisconnecting  = 4009
)

var statusText = map[int]string{
	StatusServiceRestart: "service_restart",
	StatusOK:             "ok",
	StatusBadRequest:     "bad_request",
	StatusMethodNotFound: "method_not_found",
	StatusInvalidSession: "invalid_session",
	StatusDisconnecting:  "disconnecting",
}

func StatusText(code int) string {
	return statusText[code]
}

type WebSocketMessage map[string]interface{}

type WebSocket struct {
	Conn    *websocket.Conn
	Session string
}

func (self *WebSocket) Close() {
	self.Conn.Close()
}

func (self *WebSocket) RemoteAddr() net.Addr {
	return self.Conn.RemoteAddr()
}

func (self *WebSocket) WriteJSON(message interface{}) error {
	return self.Conn.WriteJSON(message)
}

func (self *WebSocket) ReadJSON() (WebSocketMessage, error) {
	var requestMessage WebSocketMessage
	return requestMessage, self.Conn.ReadJSON(&requestMessage)
}

type WebSocketHub struct {
	api     *api.Api
	clients map[string]*WebSocket
	lock sync.RWMutex
}

func (self *WebSocketHub) Has(key string) bool {
	if nil == self.clients {
		self.clients = make(map[string]*WebSocket)
		return false
	}

	has := false
	self.lock.RLock()
	if _, ok := self.clients[key]; ok {
		has = true
	}
	self.lock.RUnlock()
	return has
}

func (self *WebSocketHub) Remove(key string) {
	self.lock.Lock()
	if _, ok := self.clients[key]; ok {
		logger.Warn("Closing web socket ", key)
		self.clients[key].Close()
		delete(self.clients, key)
	}
	self.lock.Unlock()
}

func (self *WebSocketHub) Add(key string, conn *websocket.Conn) {
	if self.Has(key) {
		self.Remove(key)
	}
	self.lock.Lock()
	wsock := &WebSocket{Conn: conn, Session: key}
	self.clients[key] = wsock
	self.lock.Unlock()
	self.listen(key, wsock)
}

func (self *WebSocketHub) listen(key string, conn *WebSocket) {
	defer hub.Remove(key)
	for {
		msg, err := conn.ReadJSON()
		if err != nil {
			logger.Error(err)
			return
		}

		b, _ := json.Marshal(msg)
		logger.Debugf(" In %v WS - %v bytes", conn.RemoteAddr(), len(b))

		resp, _ := self.api.DoJSON(string(b))
		err = conn.WriteJSON(resp)
		if err != nil {
			logger.Error(err)
			return
		}

		data, _ := resp.Marshal()
		logger.Debugf("Out %v WS - %v bytes", conn.RemoteAddr(), len(data))
	}
}

func (self *WebSocketHub) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	logger.Debugf("Opening websocket connection: %v", conn.RemoteAddr())
	if err != nil {
		logger.Error("upgrade:", err)
		return
	}

	key := uuid.New().String()
	self.Add(key, conn)
}

func New(findapi *api.Api) (*WebSocketHub, error) {
	return &WebSocketHub{
		api:     findapi,
		clients: make(map[string]*WebSocket),
	}, nil
}
