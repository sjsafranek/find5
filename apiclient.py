import requests


apiurl = 'http://localhost:8080/api/v1/find'

resp = requests.post(apiurl, json={"method":"get_users"})
