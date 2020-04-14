package main

import (
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type transmitClient struct {
	conn         *websocket.Conn
	nextChunk    []byte
	mutex        sync.Mutex
	ready        chan struct{}
	alive        bool
	shuttingDown bool
}

func newTransmitClient(conn *websocket.Conn) *transmitClient {
	t := new(transmitClient)
	t.conn = conn
	t.ready = make(chan struct{}, 1)
	return t
}

func (t *transmitClient) start() {
	go func() {
		if _, _, err := t.conn.NextReader(); err != nil {
			t.alive = false
		}
	}()
	go func() {
		t.alive = true
		for true {
			<-t.ready
			if !t.alive {
				break
			}
			t.mutex.Lock()
			c := append([]byte{}, t.nextChunk...)
			t.mutex.Unlock()
			if err := t.conn.WriteMessage(websocket.BinaryMessage, c); err != nil {
				t.alive = false
			}
		}
		log.Println("Closing websocket connection.")
		if err := t.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, "")); err == nil {
			time.Sleep(5 * time.Second)
		}
		_ = t.conn.Close()
	}()
}

func (t *transmitClient) close() {
	if t.alive {
		t.alive = false
		t.signal()
	}
}

func (t *transmitClient) signal() {
	select {
	case t.ready <- struct{}{}:
	default:
		log.Println("Dropping chunk due to slow client.")
	}
}

func (t *transmitClient) send(chunk []byte) {
	t.mutex.Lock()
	t.nextChunk = chunk
	t.mutex.Unlock()
	t.signal()
}
