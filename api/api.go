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

// ServeConnection выполняет для коннекта pre-flight checks и выступает аналогом HTTP-роутера
func ServeConnection(conn net.Conn, wg *sync.WaitGroup, st *storage.Storage) {
	for {
		// генерируем уникальный ID запроса для лога
		reqID := uniuri.New()
		le := log.WithField("request_id", reqID).WithField("remote", conn.RemoteAddr())

		// для повышения производительности нужно использовать бинарные заголовки,
		// но для экономии времени используем текстовые
		req, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Errorf("unable to read request: %s", err)
			break
		}

		fields := strings.Fields(string(req))
		if len(fields) == 0 {
			le.Errorf("empty request")
			continue
		}
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
			exp, err := setExpirationTime(st, fields[3])
			if err != nil {
				le.Errorf("wrong ttl value: %s", fields[3])
				conn.Write([]byte("wrong ttl value\n"))
				continue
			}
			// создаем объект типа "запись"
			rec := &storage.Record{
				Key:            key,
				Value:          fields[2],
				ExpirationTime: exp,
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

	switch fields[0] {
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

func setExpirationTime(st *storage.Storage, sttl string) (exp time.Time, err error) {
	ttl, err := strconv.Atoi(sttl)
	if err != nil {
		log.Errorf("error parsing request ttl: %s", err)
		return
	}

	if ttl == 0 {
		ttl = st.DefaultTTL
	}

	exp = time.Now().Add(time.Second * time.Duration(ttl))
	log.Debugf("default TTL: %d, sttl: %s, exp: %s", st.DefaultTTL, sttl, exp)
	return
}
