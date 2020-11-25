package config

import (
	"log"
	"os"
	"strconv"
)

// SrvConfig - все настройки TCP-сервера
type SrvConfig struct {
	Addr string
	Port string
}

// StorageConfig - все настройки хранилища
type StorageConfig struct {
	TTL int
}

// NewSrvConfig создает конфигурацию для конструктора сервера
func NewSrvConfig() (c SrvConfig, err error) {
	c.Addr = os.Args[1]

	_, err = strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatalf("error validating server port: %s", err)
	}
	c.Port = ":" + os.Args[2]
	return
}

// NewStorageConfig создает конфигурацию для конструктора хранилища
func NewStorageConfig() (c StorageConfig, err error) {
	c.TTL, err = strconv.Atoi(os.Args[3])
	if err != nil {
		log.Fatalf("error validating storage TTL: %s", err)
	}

	return
}
