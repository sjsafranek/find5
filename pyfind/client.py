import os
import json
import requests
from requests.auth import HTTPBasicAuth


DEFAULT_API_URL = 'http://localhost:8080'


class HttpClient(object):

    def __init__(self, username="", password="", api_url=DEFAULT_API_URL, create_user=True):
        self.api_url = api_url
        self.username = username
        self.user = None
        self.apikey = None
        self.secret = None
        if create_user:
            self.createUser(username=self.username, password=password)
        self._login(password)

    def do(self, payload):
        if 'params' not in payload:
            payload['params'] = {}
        if self.apikey:
            payload['params']['apikey'] = self.apikey
        print(json.dumps(payload))
        return requests.post("{0}/api/v1".format(self.api_url), json=payload)

    def _login(self, password):
        resp = requests.post("{0}/login?format=json".format(self.api_url), auth=HTTPBasicAuth(self.username, password))
        if 200 != resp.status_code:
            raise ValueError(resp.text)
        print(resp.text)
        respData = resp.json()
        self.user = respData['data']['user']
        self.apikey = self.user['apikey']
        self.secret = self.user['secret_token']

    def createUser(self, username="", password=""):
        return self.do({
            "method": "create_user",
            "params": {
                "username": username,
                "password": password
            }
        })

    def createDevice(self, name="", type=""):
        return self.do({
            "method": "create_device",
            "params": {
                "name": name,
                "type": type
            }
        })

    def fetchDevices(self):
        return self.do({
            "method": "get_devices"
        })

    def createSensor(self, device_id="", name="", type=""):
        return self.do({
            "method": "create_sensor",
            "params": {
                "device_id": device_id,
                "name": name,
                "type": type
            }
        })

    def fetchSensors(self, device_id=""):
        return self.do({
            "method": "get_sensors",
            "params": {
                "device_id": device_id
            }
        })

    def createLocation(self, name="", longitude=0.0, latitude=0.0):
        return self.do({
            "method": "create_location",
            "params": {
                "name": name,
                "longitude": longitude,
                "latitude": latitude
            }
        })

    def fetchLocations(self):
        return self.do({
            "method": "get_locations"
        })

    def importMeasurements(self,  device_id=None, location_id=None, data={}):
        return self.do({
            "method": "import_measurements",
            "params": {
                "location_id": location_id,
                "device_id": device_id,
                "data": data
            }
        })

    # def analyzeMeasurements(self,  device_id=None, location_id=None, data={}):
    #     return self.do({
    #         "method": "analyze",
    #         # "location_id": location_id,
    #         "device_id": device_id,
    #         "data": data
    #     })
