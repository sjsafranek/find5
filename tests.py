import os

from pyfind import database
finddb = database.Database(
    dbname = "finddb",
    dbuser = "finduser",
    dbpass = "dev"
)

username = 'admin'
location_name = 'bedroom'

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
if not devices or 0 == len([device for device in devices if device['name'] == devicename]):
    finddb.createDevice(username, devicename,'computer')
    devices = finddb.getDevices(username)

device = [device for device in devices if device['name'] == devicename][0]

# get sensor
if not device['sensors'] or 0 == len([s for s in device['sensors'] if 'wifi_card' == s['name']]):
    finddb.createSensor(device['id'], 'wifi_card', 'wifi')
    devices = finddb.getDevices(username)
    device = [device for device in devices if device['name'] == devicename][0]

sensor = [s for s in device['sensors'] if 'wifi_card' == s['name']][0]

# get location
locations = finddb.getLocations('admin')
if not locations['features'] or 0 == len([l for l in locations['features'] if l['properties']['name'] == location_name]):
    longitude = float(input("longitude: "))
    latitude = float(input("latitude: "))
    finddb.createLocation('admin', location_name, {"type":"Point","coordinates":[longitude, latitude]})
    locations = finddb.getLocations('admin')

location = [l for l in locations['features'] if l['properties']['name'] == location_name]
