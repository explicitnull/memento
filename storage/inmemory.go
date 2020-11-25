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
	Key            string
	Value          string
	ExpirationTime time.Time
}

// Storage - совукупность записей в памяти
type Storage struct {
	mtx        sync.Mutex
	data       map[string]*Record
	DefaultTTL int
	Expiration bool
}

// NewStorage создает новое хранилище
func NewStorage(conf config.StorageConfig) (st *Storage, err error) {
	if conf.TTL == 0 {
		conf.TTL = defaultTTL
	}
	st = &Storage{
		data:       make(map[string]*Record),
		DefaultTTL: conf.TTL,
		Expiration: conf.Expiration,
	}

	// запускаем удаление записей по TTL
	// TODO: наверно, стоит вынести в main для очевидности запуска
	BackgroundEvictor(st)
	return
}

// BackgroundEvictor запускает удаление устаревших записей каждые N миллисекунд.
// Если требуется точно соблюдать срок хранения, надо запускаться чаще чем раз в секунду
func BackgroundEvictor(st *Storage) {
	ticker := time.NewTicker(time.Millisecond * evictionInterval)

	go func(st *Storage) {
		for range ticker.C {
			if st.Expiration {
				st.Evict()
			}
		}
	}(st)
}

// Evict удаляет устаревшие записи
func (st *Storage) Evict() {
	start := time.Now()
	var count int

	st.mtx.Lock()
	for k, v := range st.data {
		if v.ExpirationTime.Before(start) {
			delete(st.data, k)
			count++
		}
	}
	st.mtx.Unlock()

	log.Infof("eviction done, duration: %s, records evicted: %d", time.Since(start), count)
	return
}

// ReadRecord получает запись из хранилища
func (st *Storage) ReadRecord(key string) (rec *Record, ok bool) {
	st.mtx.Lock()
	rec, ok = st.data[key]
	st.mtx.Unlock()
	log.Debugf("record: %#v", rec)
	return
}

// PutRecord кладет запись в хранилище
func (st *Storage) PutRecord(rec *Record) (err error) {
	st.mtx.Lock()
	st.data[rec.Key] = rec
	st.mtx.Unlock()
	return
}

// DeleteRecord удаляет запись из хранилища
func (st *Storage) DeleteRecord(key string) {
	st.mtx.Lock()
	delete(st.data, key)
	st.mtx.Unlock()
	return
}

const evictionInterval = 1000
