package main

import (
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

type transmitClient struct {
	conn      *websocket.Conn
	nextChunk []byte
	mutex     sync.Mutex
	ready     chan struct{}
	alive     bool
	chunkSize int
}

func newTransmitClient(conn *websocket.Conn, chunkSize int) *transmitClient {
	t := new(transmitClient)
	t.conn = conn
	t.ready = make(chan struct{}, 1)
	t.chunkSize = chunkSize
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
		for t.alive {
			<-t.ready
			t.mutex.Lock()
			c := append([]byte{}, t.nextChunk...)
			t.mutex.Unlock()
			if err := t.conn.WriteMessage(websocket.BinaryMessage, c); err != nil {
				t.alive = false
			}
		}
		_ = t.conn.Close()
	}()
}

func (t *transmitClient) send(chunk []byte) {
	t.mutex.Lock()
	t.nextChunk = chunk
	t.mutex.Unlock()
	select {
	case t.ready <- struct{}{}:
	default:
		log.Println("Dropping chunk due to slow client.")
	}
}
