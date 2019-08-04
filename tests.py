from pyfind import database
finddb = database.Database(
    dbname = "find5",
    dbuser = "findadmin",
    dbpass = "dev"
)


# finddb.createUser('sjsafranek@gmail.com','admin','dev')

# finddb.createDevice('admin','lenovo_laptop','computer')
devices = finddb.getDevices('admin')
device = devices[0]

# finddb.createLocation('admin', 'bedroom', {"type":"Point","coordinates":[-122.389664,45.434208]})

locations = finddb.getLocations('admin')
location = locations['features'][0]['properties']

# finddb.createSensor(device['id'], 'wifi_card', 'wifi')
devices = finddb.getDevices('admin')
device = devices[0]
sensor = device['sensors'][0]



print('location', location['id'])
print('sensor', sensor['id'])
