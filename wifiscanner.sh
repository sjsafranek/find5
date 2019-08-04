#!/bin/bash

SENSOR_ID=$1
LOCATION_ID=$2

while [[ true ]]; do
    sudo iwlist wlan0 scan | egrep 'SSID|Address|Signal' > tmp.txt

    python3 -c '
import re
import json
import time

from pyfind import database
finddb = database.Database(
    dbname = "find5",
    dbuser = "findadmin",
    dbpass = "dev"
)

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
            "location": "'"$LOCATION_ID"'"
        }))
        finddb.insertMeasurement("'"$SENSOR_ID"'", "'"$LOCATION_ID"'", mac_address, signal_level)
'

    sleep 10
done
