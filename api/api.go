package api

import (
	"bufio"
	"memento/storage"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dchest/uniuri"
	log "github.com/sirupsen/logrus"
)

const (
	// операции
	put    = "put"
	read   = "read"
	delete = "delete"
	bye    = "bye"
)

func HandleConnection(conn net.Conn, wg *sync.WaitGroup, st *storage.Storage) {
	for {
		req, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Errorf("unable to read request: %s", err)
			break
		}
		reqID := uniuri.New()
		le := log.WithField("request_id", reqID).WithField("remote", conn.RemoteAddr())

		fields := strings.Fields(string(req))
		operCode := fields[0]
		log.Debugf("parsed req: %v, length: %v\n", fields, len(fields))

		var resp string

		// проверяем число аргументов
		ok := validateRequest(fields)
		if !ok {
			le.Errorf("wrong %s request: %s", operCode, fields)
			conn.Write([]byte("too many or not enough parameters for this operation\n"))
			continue
		}
		key := fields[1]

		switch operCode {
		case put:
			// парсим TTL
			ttl, err := strconv.Atoi(fields[3])
			if err != nil {
				log.Errorf("error parsing request ttl: %s", err)
				return
			}
			if ttl == 0 {
				ttl = st.DefaultTTL
			}
			// создаем объект типа "запись"
			now := time.Now()
			rec := &storage.Record{
				Key:        key,
				Value:      fields[2],
				Expiration: now.Add(time.Second * time.Duration(ttl)),
			}

			resp, err = HandlePut(st, rec)
			if err != nil {
				le.Errorf("error handling request: %s: %s", fields, err)
			}
		case read:
			resp, err = HandleRead(st, key)
			if err != nil {
				le.Errorf("error handling request: %s: %s", fields, err)
			}
		case delete:
			resp, err = HandleDelete(st, key)
			if err != nil {
				le.Errorf("error handling request: %s: %s", fields, err)
			}
		case bye: // возможность для клиента завершить горутину не по таймауту, а явно
			break
		default:
			le.Errorf("unknown operation: %s", operCode)
		}
		conn.Write([]byte(resp + "\n"))
	}
	log.Debugf("closing connection")
	conn.Close()
	wg.Done()
}

func validateRequest(fields []string) (ok bool) {
	fieldsNum := len(fields)
	operCode := fields[0]

	switch operCode {
	case put:
		// формат: put key value ttl, для упрощения сделаем TTL обязательным параметром
		if fieldsNum == 4 {
			ok = true
		}
	case read, delete:
		// формат: read/delete key
		if fieldsNum == 2 {
			ok = true
		}
	}
	return
}
