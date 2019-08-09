import api

from log import NewLogger
logger = NewLogger("server")


import socket_server
socket_server.Run()


from flask import Flask, request, jsonify
app = Flask(__name__)


@app.route('/classify', methods=['POST'])
def classifyHandler():
    logger.debug('In  {0} {1}'.format(request.method, request.path))
    payload = request.get_json()
    results = api.classify(payload)
    status_code = 200
    if not results['success']:
        status_code = 400
    logger.debug('Out {0} {1} [{2}]'.format(request.method, request.path, status_code))
    return jsonify(results)


@app.route('/learn', methods=['POST'])
def learnHandler():
    logger.debug('In  {0} {1}'.format(request.method, request.path))
    payload = request.get_json()
    results = api.learn(payload)
    status_code = 200
    if not results['success']:
        status_code = 400
    logger.debug('Out {0} {1} [{2}]'.format(request.method, request.path, status_code))
    return jsonify(results)


if __name__ == "__main__":
    app.run(
        host='0.0.0.0',
        port=8002,
        debug=False)
