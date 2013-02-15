package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	mods "github.com/cwc/Gonk/modules"
	"github.com/cwc/go-v8"
	irc "github.com/fluffle/goirc/client"
)

type IModule interface {
	// Respond is called when the bot is addressed directly,
	// either in a channel (meaning that line starts with the
	// bot's nick) or via PM (meaning that target is equal to
	// the bot's nick)
	// Returns true if the module sent a response.
	Respond(target, line string, from string) bool

	// Hear is called for any message in a channel (never for PMs)
	// Returns true if the module sent a response.
	Hear(target, line string, from string) bool
}

type Module struct {
	Name    string
	Client  *irc.Conn
	Context *v8.V8Context

	respondMatchers map[*regexp.Regexp]v8.V8Function
	hearMatchers    map[*regexp.Regexp]v8.V8Function
}

func (m Module) Respond(target string, line string, from string) (responded bool) {
	// Store the target and message
	m.setTarget(target)
	m.setMessage(line, from)

	// Activate callback on any matches
	for regex, fn := range m.respondMatchers {
		count := m.setMatches(regex, line)
		if count > 0 {
			responded = true

			_, err := fn.Call(v8.V8Object{"response"})
			if err != nil {
				log.Printf("%s\n%s", err, fn)
			}
		}
	}

	return
}

func (m Module) Hear(target string, line string, from string) (responded bool) {
	// Store the target and message
	m.setTarget(target)
	m.setMessage(line, from)

	// Activate callback on any matches
	for regex, fn := range m.hearMatchers {
		count := m.setMatches(regex, line)
		if count > 0 {
			responded = true

			_, err := fn.Call(v8.V8Object{"response"})
			if err != nil {
				log.Printf("%s\n%s", err, fn)
			}
		}
	}

	return
}

func (m Module) setTarget(target string) {
	m.Context.Eval(`response.target = "` + target + `"`)
}

func (m Module) setMessage(message string, from string) {
	m.Context.Eval(`response.nick = "` + m.Client.Me.Nick + `"; response.message = {}; response.message.nick = "` + from + `"; response.message.text = "` + message + `"`)
}

func (m Module) setMatches(regex *regexp.Regexp, line string) int {
	matches := regex.FindStringSubmatch(line)
	match, _ := json.Marshal(matches)

	m.Context.Eval(`response.match = ` + string(match))

	return len(matches)
}

func (m Module) Init(script string) (ret interface{}, err error) {
	v8ctx := m.Context

	v8ctx.AddFunc("_console_log", func(args ...interface{}) interface{} {
		for _, arg := range args {
			log.Printf("> %s", arg)
		}

		return ""
	})

	v8ctx.AddFunc("_robot_respond", func(args ...interface{}) interface{} {
		regex := args[0].(*regexp.Regexp)
		fn := args[1].(v8.V8Function)

		m.respondMatchers[regex] = fn

		return ""
	})

	v8ctx.AddFunc("_robot_hear", func(args ...interface{}) interface{} {
		regex := args[0].(*regexp.Regexp)
		fn := args[1].(v8.V8Function)

		m.hearMatchers[regex] = fn

		return ""
	})

	v8ctx.AddFunc("_msg_send", func(args ...interface{}) interface{} {
		argc := len(args)

		// Last argument is expected to be the message target
		target := strings.Trim(args[argc-1].(string), `"`)

		for _, arg := range args[:argc-1] {
			text := strings.Trim(arg.(string), `"`)

			// Shorten non-embeddable URLs in the output
			_, text = mods.ShortenUrls(text, false, 0)
			m.Client.Privmsg(target, text)
		}

		return ""
	})

	v8ctx.AddFunc("_httpclient_get", func(args ...interface{}) interface{} {
		url := strings.Trim(args[0].(string), `"`)

		// Initialize client
		client := &http.Client{}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Println(err)
			return ""
		}

		// Set request headers
		var headers map[string]string
		err = json.Unmarshal([]byte(args[1].(string)), &headers)
		if err != nil {
			log.Println(err)
			return ""
		}

		for header, value := range headers {
			req.Header.Add(header, value)
		}

		// Send request
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			return ""
		}

		// Get response
		defer resp.Body.Close()
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			return ""
		}

		return string(bytes)
	})

	// Set up objects
	v8ctx.Eval(`console = {
		"log": function() { return _console_log.apply(null, arguments); }
	}`)

	v8ctx.Eval(`gonk = {
		"respond": function() { return _robot_respond.apply(null, arguments); },
		"hear" : function() { return _robot_hear.apply(null, arguments); },
	} `)

	v8ctx.Eval(`function HttpClient (url) {
		this._url = url;
		this._querystr = "";
		this._headers = {};
	}`)

	v8ctx.Eval(`HttpClient.prototype.get = function() {
		body = _httpclient_get(encodeURI(this._url + this._querystr), this._headers);
		return function(cb) {
			cb(null, null, body);
		}
	}`)

	v8ctx.Eval(`HttpClient.prototype.query = function(q) {
		prefix = "?";

		for (var prop in q) {
			this._querystr += prefix;
			this._querystr += prop + "=" + q[prop];
			prefix = "&";
		}

		return this;
	}`)

	v8ctx.Eval(`HttpClient.prototype.headers = function(headers) {
		this._headers = headers;

		return this;
	}`)

	v8ctx.Eval(`HttpClient.prototype.header = function() {
		this._headers[arguments[0]] = arguments[1];

		return this;
	}`)

	v8ctx.Eval(`response = {
		"send" : function() {
			var args = [].slice.call(arguments);
			args.push(response.target);
			_msg_send.apply(null, args); 
		},
		"random" : function(items) { return items[Math.floor(Math.random()*items.length)] },
		"http" : function(url) { return new HttpClient(url); }
	}`)

	v8ctx.Eval(`if (!String.prototype.format) {
		String.prototype.format = function() {
			var args = arguments;
			return this.replace(/{(\d+)}/g, function(match, number) {
				return typeof args[number] != 'undefined'
				? args[number]
				: match
				;
			});
		};
	}`)

	v8ctx.Eval(`module = {}`) // Module code is loaded into module.exports

	// Load script
	ret, err = v8ctx.Eval(script)
	if err != nil {
		ret = script

		return
	}

	ret, err = v8ctx.Eval(`module.exports(gonk)`)

	return
}
