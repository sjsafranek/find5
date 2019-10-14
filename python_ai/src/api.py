import os
import time
import base58
from expiringdict import ExpiringDict
from learn import AI


path_to_data = '.'

ai_cache = ExpiringDict(max_len=100000, max_age_seconds=60)


from log import NewLogger
logger = NewLogger("api")


def to_base58(family):
    return base58.b58encode(family.encode('utf-8')).decode('utf-8')

def out_file(directory, family):
    return os.path.join(directory, to_base58(family) + ".find3.ai")


def classify(payload):

    if payload is None:
        return {'success': False, 'message': 'must provide sensor data'}

    if 'sensor_data' not in payload:
        return {'success': False, 'message': 'must provide sensor data'}

    t = time.time()

    data_folder = (payload['data_folder'] if 'data_folder' in payload else DEFAULT_DATA_DIRECTORY)

    fname = out_file(data_folder, payload['sensor_data']['f'])

    ai = ai_cache.get(payload['sensor_data']['f'])
    if ai == None:
        ai = AI(to_base58(payload['sensor_data']['f']))
        logger.debug("loading {}".format(fname))
        try:
            ai.load(fname, redis_cache=True)
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

    data_folder = (payload['data_folder'] if 'data_folder' in payload else path_to_data)

    ai = AI(to_base58(payload['family']))

    # encoded file in request payload
    if 'file_data' in payload:
        ai.learn("", file_data=payload['file_data'])
    # # file on disk
    # # requires absolute path
    # elif 'csv_file' in payload:
    #     try:
    #         ai.learn( payload['csv_file'] )
    #     except FileNotFoundError:
    #         return {"success": False, "message": "could not find '{0}'".format( payload['csv_file'] )}

    fname = out_file(data_folder, payload['family'])
    ai.save(fname, redis_cache=True)

    ai_cache[payload['family']] = ai
    return {"success": True, "message": "calibrated data"}
