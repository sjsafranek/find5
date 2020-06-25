
var Utils = {};

var noop = function(){};

String.prototype.toTitleCase = function() {
    var str = this.toLowerCase().split(' ');
    for (var i = 0; i < str.length; i++) {
      str[i] = str[i].charAt(0).toUpperCase() + str[i].slice(1);
    }
    return str.join(' ');
}

Number.prototype.isInt = function() {
	return this % 1 === 0;
}

Number.prototype.isFloat = function() {
	return this % 1 !== 0;
}

Utils.md5uuid = function(len){
	return md5(Math.round(Date.now() / Math.random())).substr(0, len || 32);
}

Utils.md5object = function(val, suppress){
	if(val instanceof Object){
		if(!suppress && !Object.keys(val).length){
			console.log("Warning: msmd5 value has zero keys");
		}
		try{
			val = JSON.stringify(val);
		}catch(e){
			val = val.toString();
		}
	}
	if(!suppress && !val){
		console.log("Warning: msmd5 value is falsey");
	}
	if(!suppress && typeof val == "boolean"){
		console.log("Warning: msmd5 value is boolean");
	}
	return md5(val);
};

var FileUtils = {

    exportFile: function(content, fileName, fileType) {
        var a = document.createElement("a");
        var file = new Blob([content], {type: fileType});
        a.href = URL.createObjectURL(file);
        a.download = fileName;
        a.click();
    },

    exportJSON: function(data, fileName) {
    	var content = JSON.stringify(data)
    	FileUtils.exportFile(content, fileName, 'application/json');
    }

};
