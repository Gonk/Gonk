package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	log "github.com/fluffle/golog/logging"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"

	"github.com/Gonk/go-v8"
	irc "github.com/Gonk/goirc/client"
)

func printUsageAndExit() {
	flag.Usage()

	os.Exit(1)
}

func loadModules(conn *irc.Conn) (modules []IModule) {
	v8ctx := v8.NewContext()

	// Load each module in the modules directory
	scripts, err := ioutil.ReadDir("modules")
	if err != nil {
		log.Fatal(err.Error())
	}

	// Read base module script
	file, err := os.Open("module.js")
	if err != nil {
		log.Fatal("Error reading base module script:", err)
	}

	defer file.Close()

	baseScript, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal("Error reading base module script:", err)
	}

	for _, fileInfo := range scripts {
		if !fileInfo.IsDir() {
			if !strings.HasSuffix(fileInfo.Name(), "js") {
				continue
			}

			file, err := os.Open("modules/" + fileInfo.Name())
			if err != nil {
				log.Error("Error loading module:", err)
				continue
			}

			defer file.Close()

			script, err := ioutil.ReadAll(file)
			if err != nil {
				log.Error("Error loading module:", err)
				continue
			}

			module := NewModule(fileInfo.Name(), conn)

			// Init module with base script and its own script
			ret, err := module.Init(v8ctx, string(baseScript)+string(script))

			if err != nil {
				log.Error("Error loading module: %s\n%s", err, ret)
			}

			log.Info("Loaded module %s", fileInfo.Name())
			modules = append(modules, module)
		}
	}

	return
}

func main() {
	quitting := make(chan bool)
	disconnecting := make(chan bool)

	// Set up signal handlers
	quitSignal := make(chan os.Signal, 1)
	signal.Notify(quitSignal, os.Interrupt)
	signal.Notify(quitSignal, os.Kill)

	go func() {
		// Shutdown if quitSignal received
		<-quitSignal
		quitting <- true
	}()

	// Parse flags
	server := flag.String("server", "", "Hostname and/or port (e.g. 'localhost:6667')")
	gonkNick := flag.String("nick", "gonk", "Nickname used for the connection")
	password := flag.String("password", "", "Server password")
	ssl := flag.Bool("ssl", false, "Use SSL")
	verifyCert := flag.Bool("verify-ssl", true, "Verify SSL certificate")

	flag.Parse()

	// User must specify server
	if *server == "" {
		printUsageAndExit()
	}

	// Setup IRC client
	c := irc.SimpleClient(*gonkNick)

	if *ssl {
		c.Config().SSL = true
	}

	if !*verifyCert {
		c.Config().SSLConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// Load modules and connect them to the IRC client
	modules := loadModules(c)

	c.HandleFunc(irc.CONNECTED, func(conn *irc.Conn, line *irc.Line) {
		// Join all specified channels upon connecting
		for i := 0; i < len(flag.Args()); i++ {
			channel := fmt.Sprintf("#%s", flag.Arg(i))

			log.Info("Joining %s", channel)

			conn.Join(channel)
			conn.Privmsg(channel, "*GONK*")
		}
	})

	c.HandleFunc(irc.DISCONNECTED, func(conn *irc.Conn, line *irc.Line) {
		log.Info("Disconnected from server; shutting down")
		disconnecting <- true
	})

	c.HandleFunc("privmsg", func(conn *irc.Conn, line *irc.Line) {
		// Determine reply target
		target := line.Args[0]
		if target == conn.Me().Nick {
			// Reply via PM
			target = line.Nick
		}

		text := strings.Join(line.Args[1:], "")

		go func() {
			for _, module := range modules {
				if target == line.Nick || strings.HasPrefix(text, conn.Me().Nick) {
					// Received a PM or addressed directly in a channel
					if module.Respond(target, text, line.Nick) {
						break
					}
				} else {
					if module.Hear(target, text, line.Nick) {
						break
					}
				}
			}
		}()
	})

	err := c.ConnectTo(*server, *password)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Info(c.String())

	go func() {
		<-quitting

		c.Quit("*GONK*")

		log.Info("Shutting down")

		disconnecting <- true
	}()

	<-disconnecting
}
