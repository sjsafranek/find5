
function FindApi(user) {
	this.user = user;
}

FindApi.prototype.do = function(payload, callback) {
	payload.params = payload.params || {};
	payload.params.apikey = this.user.apikey;		// <-- always add apikey to api call
	callback = callback || console.log; 	// <-- default callback
	return $.ajax({
		method: "POST",
		url: '/api',
		contentType: "application/json",
		data: JSON.stringify(payload),
		success: function(data) {
			callback && callback(null, data.data);
		},
		error: function(jqXHR, textStatus, errorThrown) {
			console.log(jqXHR, textStatus, errorThrown);
			callback && callback(jqXHR.responseJSON || jqXHR.responseText);
		}
	});
}

FindApi.prototype.getDevices = function(callback) {
	return this.do({"method":"get_devices"}, callback);
}

FindApi.prototype.createDevice = function(name, type, callback) {
	return this.do({"method":"create_device", "params": {"name":name, "type":type}}, callback);
}

FindApi.prototype.createSensor = function(device_id, name, type, callback) {
	return this.do({"method":"create_sensor", "params": {"device_id": device_id, "name":name, "type":type}}, callback);
}

FindApi.prototype.getSensors = function(device_id, callback) {
	return this.do({"method":"get_sensors", "params": {"device_id": device_id}}, callback);
}
