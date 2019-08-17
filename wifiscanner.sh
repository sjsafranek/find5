#!/bin/bash

USERNAME=$1
PASSWORD=$2
LOCATIONNAME=$3


# setup database objects
python3 -c '
import os
import json

from pyfind import client
client = client.HttpClient(
    username="'"$USERNAME"'",
    password="'"$PASSWORD"'"
)

device_name = "{0}-{1}".format(os.getlogin(), os.uname().sysname)
client.createDevice(name=device_name, type="computer")
resp = client.fetchDevices()
if 200 != resp.status_code:
    print(resp.text)
    exit()
device = [device for device in resp.json()["data"]["devices"] if device["name"] == device_name][0]

sensor_name = "wifi_card"
client.createSensor(device_id=device["id"], name=sensor_name, type="wifi")

location_name = "'"$LOCATIONNAME"'"
client.createLocation(name=location_name)

'


# while [[ true ]]; do
for i in {1..24}; do
    sudo iwlist wlan0 scan | egrep 'SSID|Address|Signal' > wifi.tmp

    python3 -c '
import os
import re
import json
import time

from pyfind import client
client = client.HttpClient(
    username="'"$USERNAME"'",
    password="'"$PASSWORD"'"
)


device_name = "{0}-{1}".format(os.getlogin(), os.uname().sysname)
# client.createDevice(name=device_name, type="computer")
resp = client.fetchDevices()
if 200 != resp.status_code:
    print(resp.text)
    exit()
device = [device for device in resp.json()["data"]["devices"] if device["name"] == device_name][0]

sensor_name = "wifi_card"
# client.createSensor(device_id=device["id"], name=sensor_name, type="wifi")
resp = client.fetchSensors(device_id=device["id"])
if 200 != resp.status_code:
    print(resp.text)
    exit()
sensor = [sensor for sensor in resp.json()["data"]["sensors"] if sensor["name"] == sensor_name][0]

location_name = "'"$LOCATIONNAME"'"
# client.createLocation(name=location_name)
resp = client.fetchLocations()
features = resp.json()["data"]["locations"]["features"]
location = [
    feature["properties"]
        for feature in features
            if feature["properties"]["name"] == location_name
][0]


sensorMeasurements = {}

now = int(time.time())
mac_address_pattern = re.compile(r"(?:[0-9a-fA-F]:?){12}")

c = 0
with open("wifi.tmp", "r") as fh:
    lines = fh.readlines()
    for i in range(int(len(lines)/3)):
        s = i*3
        e = i*3+3
        rows = lines[s:e]
        mac_address = re.findall(mac_address_pattern, rows[0])[0]
        parts = [ item for item in rows[1].strip().split("  ") if len(item) ]
        quality = parts[0].replace("Quality=","")
        signal_level = int(parts[1].replace("Signal level=","").replace(" dBm", ""))
        essid = rows[2].split("ESSID:")[1].strip().replace("\"","")
        sensorMeasurements[mac_address] = signal_level

measurements = {}
measurements[sensor["id"]] = sensorMeasurements
resp = client.importMeasurements(device_id=device["id"], location_id=location["id"], data=measurements)
if 200 != resp.status_code:
    print(resp.text)
    exit()
print(resp.text)


'

    rm wifi.tmp
    sleep 10
done
