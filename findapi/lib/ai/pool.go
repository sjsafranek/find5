package ai

import (
	"bufio"
	"errors"
	"fmt"
	"time"

	"github.com/sjsafranek/find5/findapi/lib/ai/models"
	"github.com/sjsafranek/pool"
	"github.com/sjsafranek/logger"
)

type ClassifyPayload struct {
	Sensor models.SensorData `json:"sensor_data"`
}

const RETRY_LIMIT int = 2

func (self *AI) aiSendAndRecieveWithRetry(query string, attempt int) (string, error) {

	if RETRY_LIMIT < attempt {
		err := errors.New("retry limit reached")
		logger.Error(err)
		logger.Error(query)
		return "", err
	}

	conn, err := self.aiPool.Get()
	if nil != err {
		panic(err)
	}
	defer conn.Close()
	logger.Debug("got socket connection")

	payload := fmt.Sprintf("%v\r\n", query)
	fmt.Fprintf(conn, payload)

	results, err := bufio.NewReader(conn).ReadString('\n')
	if nil != err {
		logger.Error(err)
		attempt++
		logger.Warn("unable to read from socket")
		logger.Warn("removing socket from pool")
		pc := conn.(*pool.PoolConn)
		pc.MarkUnusable()
		pc.Close()

		// exponential backoff
		time.Sleep(time.Duration(attempt*attempt) * time.Second)

		return self.aiSendAndRecieveWithRetry(query, attempt)
	}

	// TODO
	//  - sockets get backed up and python end starts disconnecting connections
	//  - retry doesn't seem to address this
	//  - find out why sockets stop responding...
	if pc, ok := conn.(*pool.PoolConn); !ok {
		logger.Warn("socket is unusable, removing from pool")
		pc.MarkUnusable()
		pc.Close()
	}

	return results, nil
}

func (self *AI) aiSendAndRecieve(query string) (string, error) {
	// TODO
	//  - block duplicate calls
	logger.Tracef("Out  %v bytes", len(query))
	logger.Debug("sending message to ai server")

	self.guard.Lock()
	self.pending++
	self.guard.Unlock()

	results, err := self.aiSendAndRecieveWithRetry(query, 1)

	self.guard.Lock()
	self.pending--
	self.guard.Unlock()

	logger.Tracef("In %v bytes", len(results))
	logger.Tracef("In %v", results)
	return results, err
}
