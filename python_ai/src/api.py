import os
import time
import base58
from expiringdict import ExpiringDict
from learn import AI


ai_cache = ExpiringDict(max_len=100000, max_age_seconds=60)


from log import NewLogger
logger = NewLogger("api")


def to_base58(family):
    return base58.b58encode(family.encode('utf-8')).decode('utf-8')


def classify(payload):

    if payload is None:
        return {'success': False, 'message': 'must provide sensor data'}

    if 'sensor_data' not in payload:
        return {'success': False, 'message': 'must provide sensor data'}

    t = time.time()

    data_folder = '.'
    if 'data_folder' in payload:
        data_folder = payload['data_folder']

    fname = os.path.join(data_folder, to_base58(
        payload['sensor_data']['f']) + ".find3.ai")

    ai = ai_cache.get(payload['sensor_data']['f'])
    if ai == None:
        ai = AI(to_base58(payload['sensor_data']['f']), data_folder)
        logger.debug("loading {}".format(fname))
        try:
            ai.load(fname)
        except FileNotFoundError:
            logger.error('File not found')
            return {"success": False, "message": "could not find '{p}'".format(p=fname)}
        ai_cache[payload['sensor_data']['f']] = ai

    classified = ai.classify(payload['sensor_data'])

    logger.debug("classifed for {} {:d} ms".format(
        payload['sensor_data']['f'], int(1000 * (t - time.time()))))
    return {"success": True, "message": "data analyzed", 'analysis': classified}


def learn(payload):
    if payload is None:
        return {'success': False, 'message': 'must provide sensor data'}
    if 'family' not in payload:
        return {'success': False, 'message': 'must provide family'}
    if 'csv_file' not in payload and 'file_data' not in payload:
        return {'success': False, 'message': 'must provide CSV file'}
    data_folder = '.'
    if 'data_folder' in payload:
        data_folder = payload['data_folder']
    else:
        logger.debug("could not find data_folder in payload")

    logger.debug(data_folder)

    ai = AI(to_base58(payload['family']), data_folder)

    if 'file_data' in payload:
        ai.learn("", file_data=payload['file_data'])
    elif 'csv_file' in payload:
        fname = os.path.join(data_folder, payload['csv_file'])
        try:
            ai.learn(fname)
        except FileNotFoundError:
            return {"success": False, "message": "could not find '{}'".format(fname)}

    print(payload['family'])
    ai.save(os.path.join(data_folder, to_base58(
        payload['family']) + ".find3.ai"))
    ai_cache[payload['family']] = ai
    return {"success": True, "message": "calibrated data"}
