
// Geo Utils

var GeoUtils = {};

GeoUtils.getGeoLocation = function(callback) {
	callback = callback || console.log;
	if (navigator.geolocation) {
		return navigator.geolocation.getCurrentPosition(function(position) {
			callback(null, position.coords);
		});
	}
	callback(new Error("Geolocation is not supported by this browser."));
}


/**
 * Method: destinationPoint
 * Description: calculates a latLng location
 *   given an angle and distance from a starting
 *   location.
 * @param lat{float}  = Web Mercator latitude
 * @param lon{float}  = Web Mercator longitude
 * @param brng{float} = Angle from starting point
 * @param dist{float} = Distance in km
 */

Number.prototype.toRad = function() {
	return this * Math.PI / 180;
}

Number.prototype.toDeg = function() {
	return this * 180 / Math.PI;
}

GeoUtils.destinationPoint = function(lat, lon, brng, dist) {
	dist = dist / 6371;
	brng = brng.toRad();

	var lat1 = lat.toRad(), lon1 = lon.toRad();

	var lat2 = Math.asin(Math.sin(lat1) * Math.cos(dist) +
					Math.cos(lat1) * Math.sin(dist) * Math.cos(brng));

	var lon2 = lon1 + Math.atan2(Math.sin(brng) * Math.sin(dist) *
								Math.cos(lat1),
								Math.cos(dist) - Math.sin(lat1) *
								Math.sin(lat2));

	if (isNaN(lat2) || isNaN(lon2)) return null;

	return {lat: lat2.toDeg(), lon: lon2.toDeg()};
}


// Builds WKT from LeafletJS layer
// https://gist.github.com/bmcbride/4248238
GeoUtils.layerToWKT = function(layer) {
	var lng, lat, coords = [];
	if (layer instanceof L.Polygon || layer instanceof L.Polyline) {
		var latlngs = layer.getLatLngs();
	for (var i = 0; i < latlngs.length; i++) {
			var latlngs1 = latlngs[i];
			if (latlngs1.length){
			for (var j = 0; j < latlngs1.length; j++) {
				coords.push(latlngs1[j].lng + " " + latlngs1[j].lat);
				if (j === 0) {
					lng = latlngs1[j].lng;
					lat = latlngs1[j].lat;
				}
			}}
			else
			{
				coords.push(latlngs[i].lng + " " + latlngs[i].lat);
				if (i === 0) {
					lng = latlngs[i].lng;
					lat = latlngs[i].lat;
				}}
	};
		if (layer instanceof L.Polygon) {
			return "POLYGON((" + coords.join(",") + "," + lng + " " + lat + "))";
		} else if (layer instanceof L.Polyline) {
			return "LINESTRING(" + coords.join(",") + ")";
		}
	} else if (layer instanceof L.Marker) {
		return "POINT(" + layer.getLatLng().lng + " " + layer.getLatLng().lat + ")";
	}
};


// Converts numeric bearing to human readable cardinal direction
GeoUtils.bearingToCardinalDirection = function(bearing) {
	var bearingword = '';
	if      (bearing >=  22 && bearing <=  67) bearingword = 'North East';
	else if (bearing >=  67 && bearing <= 112) bearingword =  'East';
	else if (bearing >= 112 && bearing <= 157) bearingword = 'South East';
	else if (bearing >= 157 && bearing <= 202) bearingword =  'South';
	else if (bearing >= 202 && bearing <= 247) bearingword = 'South West';
	else if (bearing >= 247 && bearing <= 292) bearingword =  'West';
	else if (bearing >= 292 && bearing <= 337) bearingword = 'North West';
	else if (bearing >= 337 || bearing <=  22) bearingword =  'North';
	return bearingword;
};

GeoUtils.geocode = function(placename, callback) {
	$.ajax({
		method: 'GET',
		url: 'https://nominatim.openstreetmap.org/search',
		data: {
			q: placename,
			format: 'json'
		},
		success: function(data) {
			callback(null, data[0]);
		}
	});
}

GeoUtils.reverseGeocode = function(coords, callback) {
	$.ajax({
		method: 'GET',
		url: 'https://nominatim.openstreetmap.org/reverse',
		data: {
			lat: coords.latitude,
			lon: coords.longitude,
			format: 'json'
		},
		success: function(data) {
			callback(null, data);
		}
	});
}
