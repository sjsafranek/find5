import os
import json
import time
import signal
import socket
import traceback
import threading
# import _thread

import api

from log import NewLogger
logger = NewLogger("tcp")

# tcp server
# TCP_IP = '127.0.0.1'
TCP_IP = 'localhost'
TCP_PORT = 7005
BUFFER_SIZE = 1024


def on_new_client(conn, addr):
    try:
        payload = ''
        while True:

            logger.debug("pipe contains {0} bytes".format(len(payload)))

            chunk = conn.recv(BUFFER_SIZE).decode()
            if not chunk:
                break

            logger.debug("chunk recieved {0} bytes".format(len(chunk)))
            payload += chunk
            if not "\n" in payload:
                logger.debug("new line not detected in chunk. waiting for more chunks")
                continue

            parts = payload.split("\n")
            payload = parts[1]

            if parts[0].count('{') == parts[0].count('}') and 0 != parts[0].count('{'):

                query = json.loads(parts[0])
                logger.info("IN  {0} bytes".format(len(json.dumps(query))))

                results = {"success": False, "message": "incorrect usage"}
                if 'method' in query and 'data' in query:
                    if 'learn' == query['method']:
                        results = api.learn(query['data'])
                    elif 'classify' == query['method']:
                        results = api.classify(query['data'])
                elif 'get_cache' == query['method']:
                    results = api.ai_cache

                logger.info("OUT {0}".format(json.dumps(results)))
                response = json.dumps(results)
                conn.send("{0}\n".format(response).encode())

    except Exception as e:
        logger.error(e)
        with open("__query.json","w") as fh:
            fh.write(parts[0])
        traceback.print_exc()

    logger.warn("client closed socket")
    conn.close()


class TcpServer(threading.Thread):
    def __init__(self):
        threading.Thread.__init__(self)
        self.event = threading.Event()

    def run(self):
        try:
            logger.info("starting up on {0} port {1}".format(TCP_IP, TCP_PORT))
            self.socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self.socket.bind((TCP_IP, TCP_PORT))
            self.socket.listen(1)
            self._acceptClients()
        except Exception as e:
            logger.error(e)
            # fatal error occured
            # bail main process
            os._exit(1)

    def _acceptClients(self):
        logger.debug('waiting for connections')
        while not self.event.is_set() :
            conn, addr = self.socket.accept()
            logger.debug('Connected address: {0}'.format(addr))
            # _thread.start_new_thread(on_new_client, (conn, addr))
            t = threading.Thread(target=on_new_client, args=(conn, addr,))
            t.start()
            # try:
            #     payload = ''
            #     while True:
            #
            #         logger.debug("pipe contains {0} bytes".format(len(payload)))
            #
            #         chunk = conn.recv(BUFFER_SIZE).decode()
            #         if not chunk:
            #             logger.warn("client closed socket")
            #             break
            #
            #         logger.debug("chunk recieved {0} bytes".format(len(chunk)))
            #         payload += chunk
            #         if not "\n" in payload:
            #             logger.debug("new line not detected in chunk. waiting for more chunks")
            #             continue
            #
            #         parts = payload.split("\n")
            #         payload = parts[1]
            #
            #         if parts[0].count('{') == parts[0].count('}') and 0 != parts[0].count('{'):
            #
            #             query = json.loads(parts[0])
            #             logger.info("IN  {0}".format(json.dumps(query)))
            #
            #             results = {"success": False, "message": "incorrect usage"}
            #             if 'method' in query and 'data' in query:
            #                 if 'learn' == query['method']:
            #                     results = api.learn(query['data'])
            #                 elif 'classify' == query['method']:
            #                     results = api.classify(query['data'])
            #             elif 'get_cache' == query['method']:
            #                 results = api.ai_cache
            #
            #             logger.info("OUT {0}".format(json.dumps(results)))
            #             response = json.dumps(results)
            #             conn.send("{0}\n".format(response).encode())
            #
            # except Exception as e:
            #     logger.error(e)
            #
            # conn.close()

    def shutdown(self):
        self.event.set()
        self.socket.close()
        raise ValueError("SHUTDOWN")


def Run():
    thread = TcpServer()
    thread.setDaemon(True)
    thread.start()
