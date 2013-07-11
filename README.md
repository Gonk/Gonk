Gonk
====

Your friendly neighborhood IRC droid

## What is Gonk?

Gonk is an IRC bot written in [Go](http://golang.org). Its modular design and embedded JavaScript engine make it easy to dynamically add functionality via "modules".

The API exposed to modules replicates that of [Hubot](http://github.com/github/hubot), so that scripts for Hubot can be used interchangeably with Gonk with few or no modifications.

Gonk does not currently support CoffeeScript - scripts must be translated to JavaScript before they can be used.

## Usage

```
$ Gonk -server=irc.host.com -ssl=true -password=serverPassword channel1 channel2
```

On startup, Gonk will search the `modules` directory and attempt to load any file with the extension `.js` as a module.

## Dependencies

* Go 1.1+
* Make/GCC
* [Google V8](https://code.google.com/p/v8/) headers and library

## How to Build

Gonk embeds the V8 engine and binds to it with [go-v8](http://github.com/Gonk/go-v8). This complicates the build process a bit. These instructions assume that a proper compiler toolchain and the V8 library and header files exist on your system in the standard location and that you understand building with Makefiles and the Go workflow (e.g. how to use `GOPATH`).

---

First, Gonk and its dependencies need to be installed into your `GOPATH`:

```
$ go get github.com/Gonk/Gonk
```

**This command will appear to fail for go-v8.** This is because go-v8 requires an additional build step:

```
$ cd $GOPATH/src/github.com/Gonk/go-v8
$ make
```

This will build a V8 wrapper library and install it in your `GOPATH`. You should now be able to finish building and run Gonk:

```
$ go get github.com/Gonk/Gonk
$ $GOPATH/bin/Gonk
```
