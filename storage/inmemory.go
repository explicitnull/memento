package storage

import (
	"memento/config"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const defaultTTL = 3600

// Record - единица хранения БД
type Record struct {
	Key        string
	Value      string
	Expiration time.Time
}

// Storage - совукупность записей в памяти
type Storage struct {
	mtx        sync.Mutex
	data       map[string]*Record
	DefaultTTL int
}

// NewStorage создает новое хранилище
func NewStorage(conf config.StorageConfig) (st *Storage, err error) {
	if conf.TTL == 0 {
		conf.TTL = defaultTTL
	}
	st = &Storage{
		data:       make(map[string]*Record),
		DefaultTTL: conf.TTL,
	}
	return
}

func (st *Storage) ReadRecord(key string) (rec *Record, ok bool) {
	st.mtx.Lock()
	rec, ok = st.data[key]
	st.mtx.Unlock()
	log.Debugf("record: %#v", rec)
	return
}

func (st *Storage) PutRecord(rec *Record) (err error) {
	st.mtx.Lock()
	st.data[rec.Key] = rec
	st.mtx.Unlock()
	return
}

func (st *Storage) DeleteRecord(key string) {
	st.mtx.Lock()
	delete(st.data, key)
	st.mtx.Unlock()
	return
}
