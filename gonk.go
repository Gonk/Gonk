package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	irc "github.com/fluffle/goirc/client"
)

func printUsageAndExit() {
	flag.Usage()

	os.Exit(1)
}

func main() {
	quitting := make(chan bool)
	disconnecting := make(chan bool)

	// Set up ^C handler
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		// Quit on ^C
		<-interrupt
		quitting <- true
	}()

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

	c := irc.SimpleClient(*gonkNick)

	if *ssl {
		c.SSL = true
	}

	if !*verifyCert {
		c.SSLConfig = &tls.Config{InsecureSkipVerify: true}
	}

	c.AddHandler("connected", func(conn *irc.Conn, line *irc.Line) {
		// Join all specified channels upon connecting
		for i := 0; i < len(flag.Args()); i++ {
			channel := fmt.Sprintf("#%s", flag.Arg(i))

			log.Printf("Joining %s", channel)

			conn.Join(channel)
			conn.Privmsg(channel, "*GONK*")
		}
	})

	c.AddHandler("disconnected", func(conn *irc.Conn, line *irc.Line) {
		disconnecting <- true
	})

	c.AddHandler("privmsg", func(conn *irc.Conn, line *irc.Line) {
		// Determine reply target
		target := line.Args[0]
		if target == *gonkNick {
			target = line.Nick
		}

		// Send reply
		c.Privmsg(target, "*GONK*")
	})

	err := c.Connect(*server, *password)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		<-quitting

		c.Quit("*GONK*")

		disconnecting <- true
	}()

	<-disconnecting
}
