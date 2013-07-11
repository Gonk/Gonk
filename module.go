package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	log "github.com/fluffle/golog/logging"
	"net/http"
	"regexp"
	"strings"

	mods "github.com/Gonk/Gonk/modules"
	"github.com/Gonk/go-v8"
	irc "github.com/Gonk/goirc/client"
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
				log.Error("%s\n%s", err, fn)
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
				log.Error("%s\n%s", err, fn)
			}
		}
	}

	return
}

func (m Module) setTarget(target string) {
	m.Context.Eval(`response.target = "` + target + `"`)
}

func (m Module) setMessage(message string, from string) {
	m.Context.Eval(`response.nick = "` + m.Client.Me().Nick + `"; response.message = {}; response.message.nick = "` + from + `"; response.message.text = "` + message + `"`)
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
			log.Warn("> %s", arg)
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

			// Shorten URLs in the response
			_, text = mods.ShortenUrls(text, false, true, 25)

			m.Client.Privmsg(target, text)
		}

		return ""
	})

	v8ctx.AddFunc("_httpclient_post", func(args ...interface{}) interface{} {
		url := strings.Trim(args[0].(string), `"`)

		// Initialize client
		client := &http.Client{}

		req, err := http.NewRequest("POST", url, bytes.NewBufferString(strings.Trim(args[2].(string), `"`)[1:]))
		if err != nil {
			log.Error(err.Error())
			return ""
		}

		// Set request headers
		var headers map[string]string
		err = json.Unmarshal([]byte(args[1].(string)), &headers)
		if err != nil {
			log.Error(err.Error())
			return ""
		}

		for header, value := range headers {
			req.Header.Add(header, value)
		}

		// Send request
		resp, err := client.Do(req)
		if err != nil {
			log.Error(err.Error())
			return ""
		}

		// Get response
		defer resp.Body.Close()
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err.Error())
			return ""
		}

		return string(bytes)
	})

	v8ctx.AddFunc("_httpclient_get", func(args ...interface{}) interface{} {
		url := strings.Trim(args[0].(string), `"`)

		// Initialize client
		client := &http.Client{}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Error(err.Error())
			return ""
		}

		// Set request headers
		var headers map[string]string
		err = json.Unmarshal([]byte(args[1].(string)), &headers)
		if err != nil {
			log.Error(err.Error())
			return ""
		}

		for header, value := range headers {
			req.Header.Add(header, value)
		}

		// Send request
		resp, err := client.Do(req)
		if err != nil {
			log.Error(err.Error())
			return ""
		}

		// Get response
		defer resp.Body.Close()
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err.Error())
			return ""
		}

		return string(bytes)
	})

	// Load script
	ret, err = v8ctx.Eval(script)
	if err != nil {
		ret = script

		return
	}

	ret, err = v8ctx.Eval(`module.exports(gonk)`)

	return
}
