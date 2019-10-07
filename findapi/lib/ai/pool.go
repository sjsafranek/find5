package ai

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/sjsafranek/find5/findapi/lib/ai/models"
	"github.com/sjsafranek/pool"
)

var (
	AI_SERVER_ADDRESS string = "localhost:7005"
	AI_POOL           pool.Pool
	AI_PENDING        int = 0
	ai_counter_lock   sync.RWMutex
)

func init() {
	factory := func() (net.Conn, error) { return net.Dial("tcp", AI_SERVER_ADDRESS) }
	go func() {
		for {
			pool, err := pool.NewChannelPool(4, 10, factory)
			if nil != err {
				logger.Warn("Unable to communicate with AI server")
				time.Sleep(5 * time.Second)
				continue
			}
			logger.Info("Connected to AI server")
			AI_POOL = pool
			break
		}
	}()
}

type ClassifyPayload struct {
	Sensor models.SensorData `json:"sensor_data"`
	// DataFolder string            `json:"data_folder"`
}

const RETRY_LIMIT int = 2

func aiSendAndRecieveWithRetry(query string, attempt int) (string, error) {
	// logger.Debug(query)

	if RETRY_LIMIT < attempt {
		err := errors.New("retry limit reached")
		logger.Error(err)
		logger.Error(query)
		return "", err
	}

	conn, err := AI_POOL.Get()
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

		return aiSendAndRecieveWithRetry(query, attempt)
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

func aiSendAndRecieve(query string) (string, error) {
	// TODO
	//  - block duplicate calls
	logger.Tracef("Out  %v bytes", len(query))
	logger.Debug("sending message to ai server")

	ai_counter_lock.Lock()
	AI_PENDING++
	ai_counter_lock.Unlock()

	results, err := aiSendAndRecieveWithRetry(query, 1)

	ai_counter_lock.Lock()
	AI_PENDING--
	ai_counter_lock.Unlock()

	logger.Tracef("In %v bytes", len(results))
	logger.Tracef("In %v", results)
	return results, err
}

func init() {

	go func() {
		for {
			time.Sleep(10 * time.Second)
			ai_counter_lock.RLock()
			if 0 != AI_PENDING {
				logger.Debugf("%v pending AI requests", AI_PENDING)
			}
			ai_counter_lock.RUnlock()
		}
	}()

}

func Shutdown() {
	logger.Warn("Closing connection pool...")
	AI_POOL.Close()
}
