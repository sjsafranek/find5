#!/bin/bash

# SENSOR_ID=$1
# LOCATION_ID=$2

USERNAME=$1
LOCATIONNAME=$2

while [[ true ]]; do
    sudo iwlist wlan0 scan | egrep 'SSID|Address|Signal' > tmp.txt

    python3 -c '
import os
import re
import json
import time

from pyfind import database
finddb = database.Database(
    dbname = "finddb",
    dbuser = "finduser",
    dbpass = "dev"
)


username = "'"$USERNAME"'"
location_name = "'"$LOCATIONNAME"'"

# get user
user = finddb.getUser(username)
if not user:
    print("Creating {0} username".format(username))
    email = input("email: ")
    while True:
        password1 = input("password: ")
        password2 = input("password(again): ")
        if password1 == password2:
            break
        print("passwords do not match")
    finddb.createUser(email, username, password1)
    user = finddb.getUser(username)

user = user[0]

# get device
devicename = "{0}-{1}".format(os.getlogin(), os.uname().sysname)
devices = finddb.getDevices(username)
if not devices or 0 == len([device for device in devices if device["name"] == devicename]):
    finddb.createDevice(username, devicename,"computer")
    devices = finddb.getDevices(username)

device = [device for device in devices if device["name"] == devicename][0]

# get sensor
if not device["sensors"] or 0 == len([s for s in device["sensors"] if "wifi_card" == s["name"]]):
    finddb.createSensor(device["id"], "wifi_card", "wifi")
    devices = finddb.getDevices(username)
    device = [device for device in devices if device["name"] == devicename][0]

sensor = [s for s in device["sensors"] if "wifi_card" == s["name"]][0]
print(sensor)

# get location
locations = finddb.getLocations("admin")
if not locations["features"] or 0 == len([l for l in locations["features"] if l["properties"]["name"] == location_name]):
    longitude = float(input("longitude: "))
    latitude = float(input("latitude: "))
    finddb.createLocation("admin", location_name, {"type":"Point","coordinates":[longitude, latitude]})
    locations = finddb.getLocations("admin")

location = [l for l in locations["features"] if l["properties"]["name"] == location_name][0]



now = int(time.time())
mac_address_pattern = re.compile(r"(?:[0-9a-fA-F]:?){12}")

c = 0
with open("tmp.txt", "r") as fh:
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
        print(json.dumps({
            "event_timestamp": now,
            "mac_address": mac_address,
            "essid": essid,
            "quality": quality,
            "signal_level_dBm": signal_level,
            "location_id": location["properties"]["id"],
            "sensor_id": sensor["id"]
        }))
        finddb.insertMeasurement(sensor["id"], location["properties"]["id"], mac_address, signal_level)
'

    sleep 10
done
