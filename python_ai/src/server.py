import os
import json
import time
import signal
import socket
import argparse
import traceback
import threading


import api

from log import NewLogger
logger = NewLogger("server")


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
                # debugging method
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





if '__main__' == __name__:

    parser = argparse.ArgumentParser(description="Machine Learning AI for FIND")
    parser.add_argument ('-p', '--port', type=int, help='port', default=7005)
    parser.add_argument ('--host', type=str, help='host', default='localhost')
    parser.add_argument ('-D', '--data_directory', type=str, help='data directory', default='.')
    args = parser.parse_args()

    api.DEFAULT_DATA_DIRECTORY = args.data_directory

    try:
        logger.info("starting up on {0} port {1}".format(args.host, args.port))
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.bind((args.host, args.port))
        sock.listen(1)

        logger.debug('waiting for connections')
        while True:
            conn, addr = sock.accept()
            logger.debug('Connected address: {0}'.format(addr))
            t = threading.Thread(target=on_new_client, args=(conn, addr,))
            t.start()

    except Exception as e:
        logger.error(e)
