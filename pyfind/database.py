import json
import psycopg2

# dbname = 'find5'
# dbuser = 'findadmin'
# dbpass = 'dev'

# connection = psycopg2.connect('dbname={0} user={1} password={2}'.format(dbname, dbuser, dbpass))
# cursor = connection.cursor()


class Database(object):
    def __init__(self, dbname='', dbuser='', dbpass=''):
        self._connection = psycopg2.connect('dbname={0} user={1} password={2}'.format(dbname, dbuser, dbpass))
        self._cursor = self._connection.cursor()

    def _insert(self, query, args):
        try:
            self._cursor.execute(
                query,
                args
            )
            self._connection.commit()
        except Exception as e:
            print(e)
            self._connection.rollback()

    def createUser(self, email, username, password):
        self._insert("""
            INSERT INTO users (email, username, password)
                VALUES (%s, %s, %s)""",
            (email, username, password)
        )

    def getUser(self, username):
        self._cursor.execute("""
            SELECT json_agg(u) FROM (
                SELECT
                    username,
                    email,
                    apikey,
                    secret_token,
                    is_deleted,
                    created_at,
                    updated_at
                FROM users
                WHERE username = %s
            ) u;
        """, (username,) )
        return self._cursor.fetchone()[0]

    def createDevice(self, username, name, dtype):
        self._insert("""
            INSERT INTO devices(username, name, type)
                VALUES (%s, %s, %s)""",
            (username, name, dtype)
        )

    def getDevices(self, username):
        self._cursor.execute("""
            SELECT json_agg(d) FROM (
                SELECT
                    devices.*,
                    (
                        SELECT json_agg(s) FROM (
                            SELECT
                                sensors.*
                            FROM sensors
                            WHERE device_id=devices.id
                        ) s
                    ) AS sensors
                FROM devices
                WHERE username = %s
            ) d;
        """, (username,) )
        return self._cursor.fetchone()[0]

    def createLocation(self, username, name, geojson):
        self._insert("""
            INSERT INTO locations (username, name, geom)
                VALUES (%s, %s, ST_SetSRID(ST_GeomFromGeoJSON(%s::TEXT),4326));
            """,
            (username, name, json.dumps(geojson))
        )

    def getLocations(self, username):
        self._cursor.execute("""
            SELECT
                json_build_object(
                    'type', 'FeatureCollection',
                    'features', json_agg(c)
                ) AS geojson
            FROM (
                SELECT
                    'Feature' AS type,
                    ST_AsGeoJSON(geom)::jsonb AS geometry,
                    json_build_object(
                        'id', id,
                        'name', name,
                        'username', username,
                        'created_at', created_at,
                        'updated_at', updated_at
                    ) AS properties
                FROM locations
                WHERE
                        geom IS NOT NULL
                    AND
                        username = %s
            ) c;
        """, (username,) )
        return self._cursor.fetchone()[0]

    def createSensor(self, device_id, name, stype):
        self._insert("""
            INSERT INTO sensors (device_id, name, type)
                VALUES (%s, %s, %s)""",
            (device_id, name, stype)
        )

    def insertMeasurement(self, sensor_id, location_id, key, value):
        self._insert("""
            INSERT INTO measurements (sensor_id, location_id, key, value)
                VALUES (%s, %s, %s, %s)""",
            (sensor_id, location_id, key, value)
        )