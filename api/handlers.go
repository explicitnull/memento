package api

import (
	"memento/storage"

	log "github.com/sirupsen/logrus"
)

// Возвращает запись по ключу
func HandleRead(st *storage.Storage, key string) (resp string, err error) {
	rec, _ := st.ReadRecord(key)
	if rec.Value != "" {
		// format: "value" "expiration"
		resp = rec.Value + " " + rec.Expiration.Format("RFC3339")
	} else {
		resp = "record not found"
	}
	log.WithField("key", key).WithField("value", rec.Value).Infof("read request handled")
	return
}

// Записывает/перезаписывает запись по ключу
func HandlePut(st *storage.Storage, rec *storage.Record) (resp string, err error) {
	st.PutRecord(rec)
	resp = "record put"
	log.WithField("key", rec.Key).WithField("value", rec.Value).Infof("put request handled")
	return
}

// HandleDelete удаляет запись по ключу
func HandleDelete(st *storage.Storage, key string) (resp string, err error) {
	// сначала узнаем, есть ли запись
	rec, _ := st.ReadRecord(key)
	if rec.Value == "" {
		resp = "record not found"
	} else {
		st.DeleteRecord(key)
		resp = "record removed"
	}
	return
}
