import requests
from requests.auth import HTTPBasicAuth

baseUrl = 'http://localhost:8080'

resp = requests.post(baseUrl+"/api/v1/find", json={"method":"create_user","username":"admin","email":"admin@email.com","password":"test"})
print(resp.text)

resp = requests.post(baseUrl+"/login", auth=HTTPBasicAuth("admin","test"))
print(resp.text)
