
var Config = function(data) {
    this.data = data;
}

Config.prototype.get = function(key) {
    return JSON.parse(JSON.stringify(this.data[key]));
}
