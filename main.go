package main

import (
	"fmt"
	"net"
	"os"
	"sync"

	"memento/api"
	"memento/config"
	"memento/storage"

	log "github.com/sirupsen/logrus"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: memento address port ttl")
		os.Exit(1)
	}

	log.SetLevel(log.DebugLevel)

	// создаем хранилище
	stConfig, err := config.NewStorageConfig()
	if err != nil {
		log.Fatalf("error creating storage config: %s", err)
	}

	primStorage, err := storage.NewStorage(stConfig)
	if err != nil {
		log.Fatalf("error creating storage: %s", err)
	}

	// создаем TCP-сервер
	srvConfig, err := config.NewSrvConfig()
	if err != nil {
		log.Fatalf("error creating server config: %s", err)
	}

	listener, err := net.Listen("tcp", srvConfig.Addr+string(srvConfig.Port))
	if err != nil {
		log.Fatalf("unable to listen: %s", err)
	}

	var wg sync.WaitGroup

	// принимаем коннекты
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("error accepting: %s", err)
		}
		log.Debugf("accepted connection from %s", conn.RemoteAddr())

		wg.Add(1)
		go api.HandleConnection(conn, &wg, primStorage)
	}

}
