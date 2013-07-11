// This script defines objects loaded into the context of every Gonk module

console = {
	"log": function() { return _console_log.apply(null, arguments); }
}

gonk = {
	"respond": function() { return _robot_respond.apply(null, arguments); },
	"hear" : function() { return _robot_hear.apply(null, arguments); },
} 

function HttpClient (url) {
	this._url = url;
	this._querystr = "";
	this._headers = {};
}

HttpClient.prototype.post = function() {
	uri = encodeURI(this._url);
	body = _httpclient_post(uri, this._headers, this._querystr);
	return function(cb) {
		cb(null, null, body);
	}
}

HttpClient.prototype.get = function() {
	uri = encodeURI(this._url) + this._querystr;
	body = _httpclient_get(uri, this._headers);
	return function(cb) {
		cb(null, null, body);
	}
}

HttpClient.prototype.query = function(q) {
	prefix = "?";

	for (var prop in q) {
		this._querystr += prefix;
		this._querystr += prop + "=" + encodeURIComponent(q[prop]);
		prefix = "&";
	}

	return this;
}

HttpClient.prototype.headers = function(headers) {
	this._headers = headers;

	return this;
}

HttpClient.prototype.header = function() {
	this._headers[arguments[0]] = arguments[1];

	return this;
}

response = {
	"send" : function() {
		var args = [].slice.call(arguments);
		args.push(response.target);
		_msg_send.apply(null, args); 
	},
	"random" : function(items) { return items[Math.floor(Math.random()*items.length)] },
	"http" : function(url) { return new HttpClient(url); }
}

if (!String.prototype.format) {
	String.prototype.format = function() {
		var args = arguments;
		return this.replace(/{(\d+)}/g, function(match, number) {
			return typeof args[number] != 'undefined'
			? args[number]
			: match
			;
		});
	};
}

module = {}; // Module code is loaded into module.exports
