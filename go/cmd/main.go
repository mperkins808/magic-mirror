package main

import (
	"flag"
	"os"

	"github.com/mperkins808/magic-mirror/go/pkg/magicmirror"
	log "github.com/sirupsen/logrus"
)

var (
	name   string
	remote string
	local  string
	apikey string
	h      bool
)

func main() {
	parseFlags()

}

func parseFlags() {
	flag.StringVar(&name, "name", "", "optionally name your connection")
	flag.StringVar(&remote, "remote", "", "the remote host to connect to. Must be a websocket accepting endpoint")
	flag.StringVar(&local, "local", "", "optionally restrict all connections to a localhost. eg http://localhost:9090")
	flag.StringVar(&apikey, "apikey", "", "supply an api key if required")
	flag.BoolVar(&h, "help", false, "display help information")
	flag.Parse()

	if h {
		help()
		return
	}

	if remote == "" {
		help()
		log.Fatal("remote host must be specified")
	}

	if os.Getenv("SAFE") == "true" {
		if local == "" {
			help()
			log.Fatal("because safe mode is active. you must supply a localhost to forward requests to")
		}
	}

	magicmirror.Mirror(remote, local, name, apikey)
}

func help() {
	log.Println("usage of magicmirror\n")
	flag.PrintDefaults()
}
