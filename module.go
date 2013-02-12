package main

import (
	"encoding/json"
	"log"
	"regexp"
	"strings"

	"github.com/cwc/go-v8"
	irc "github.com/fluffle/goirc/client"
)

type IModule interface {
	// Respond is called when the bot is addressed directly,
	// either in a channel (meaning that line starts with the
	// bot's nick) or via PM (meaning that target is equal to
	// the bot's nick)
	Respond(target, line string)

	// Hear is called for any message in a channel (never for PMs)
	Hear(target, line string)
}

type Module struct {
	Name    string
	Client  *irc.Conn
	Context *v8.V8Context

	respondMatchers map[*regexp.Regexp]v8.V8Function
	hearMatchers    map[*regexp.Regexp]v8.V8Function
}

func (m Module) Respond(target string, line string) {
	// Store the target
	m.setTarget(target)

	// Activate callback on any matches
	for regex, fn := range m.respondMatchers {
		count := m.setMatches(regex, line)
		if count > 0 {
			fn.Call(`response`)
		}
	}
}

func (m Module) Hear(target string, line string) {
	// Store the target
	m.setTarget(target)

	// Activate callback on any matches
	for regex, fn := range m.hearMatchers {
		count := m.setMatches(regex, line)
		if count > 0 {
			fn.Call(`response`)
		}
	}
}

func (m Module) setTarget(target string) {
	m.Context.Eval(`response.target = "` + target + `"`)
}

func (m Module) setMatches(regex *regexp.Regexp, line string) int {
	matches := regex.FindStringSubmatch(line)
	match, _ := json.Marshal(matches)

	m.Context.Eval(`response.match = ` + string(match))

	return len(match)
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
		target := strings.Trim(args[1].(string), `"`)
		text := strings.Trim(args[0].(string), `"`)

		m.Client.Privmsg(target, text)

		return ""
	})

	// Set up objects
	v8ctx.Eval(`console = {
                "log": function() { return _console_log.apply(null, arguments); }
    }`)

	v8ctx.Eval(`gonk = {
                "respond": function() { return _robot_respond.apply(null, arguments); },
                "hear" : function() { return _robot_hear.apply(null, arguments); }
    } `)

	v8ctx.Eval(`response = {
                "send" : function() {
                    var args = [].slice.call(arguments);
                    args.push(response.target);
                    _msg_send.apply(null, args); 
                }
    }`)

	v8ctx.Eval(`module = {}`) // Module code is loaded into module.exports

	// Load script
	ret, err = v8ctx.Eval(script)
	v8ctx.Eval(`module.exports(gonk)`)

	return
}
