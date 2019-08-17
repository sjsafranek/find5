import json
import time

from pyfind import client
client = client.HttpClient(
    username="admin",
    password="test1234"
)


device_name = "zacks_test_data"
client.createDevice(name=device_name, type="test_data")
resp = client.fetchDevices()
if 200 != resp.status_code:
    print(resp.text)
    exit()
device = [device for device in resp.json()["data"]["devices"] if device["name"] == device_name][0]

sensor_name = "wifi_card"
client.createSensor(device_id=device["id"], name=sensor_name, type="wifi")

resp = client.fetchSensors(device_id=device["id"])
if 200 != resp.status_code:
    print(resp.text)
    exit()
sensor = [sensor for sensor in resp.json()["data"]["sensors"] if sensor["name"] == sensor_name][0]

print(device['id'], sensor['id'])


testLocations = {}
with open('testing/testdb.learn.1439597065993.jsons_old', 'r') as f:
    for line in f.readlines():
        data = json.loads(line)
        testLocations[data['l']] = ''

for location_name in testLocations:
    client.createLocation(name=location_name)

resp = client.fetchLocations()
features = resp.json()["data"]["locations"]["features"]

for feature in features:
    if feature["properties"]["name"] in testLocations:
        testLocations[feature["properties"]["name"]] = feature["properties"]["id"]


# run tests
startTime = time.time()
with open('testing/testdb.learn.1439597065993.jsons_old', 'r') as f:
    for line in f.readlines():
        data = json.loads(line)
        resp = client.importMeasurements(
            device_id=device['id'],
            location_id=testLocations[data['l']],
            data={
                sensor['id']: data['s']['wifi']
            }
        )
        print(resp)
print("run time: ",time.time()-startTime)
