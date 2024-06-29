package main

import (
	"flag"
	"log"
)

func main() {
	configPath := flag.String("cp", "config.json", "-cp=/path/to/config.json")

	flag.Parse()
	log.SetFlags(0)

	Rlog.Infof("loading config from %s", *configPath)
	conf, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatal(err)
		return
	}

	api := newAPI(conf)
	go api.ConfigureRoutes()

	err = initWebsocketClient(conf)
	if err != nil {
		Rlog.Fatal("initWebsocketClient error: ", err)
	}
}
