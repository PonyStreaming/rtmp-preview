package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"path"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type streamer struct {
	transcoders map[string]*transcoder
	baseURL     string
	chunkSize   int

	resolution   string
	videoBitrate string
	audioBitrate string
}

func makeStreamer(baseURL string, chunkSize int, resolution, videoBitrate, audioBitrate string) streamer {
	return streamer{
		transcoders:  map[string]*transcoder{},
		baseURL:      baseURL,
		chunkSize:    chunkSize,
		resolution:   resolution,
		videoBitrate: videoBitrate,
		audioBitrate: audioBitrate,
	}
}

func (s *streamer) handle(w http.ResponseWriter, r *http.Request) {
	_, streamName := path.Split(r.URL.Path)
	if streamName == "" {
		http.Error(w, "Not found", http.StatusNotFound)
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrading connection to websocket failed: %v", err)
		return
	}
	if _, ok := s.transcoders[streamName]; !ok {
		s.transcoders[streamName] = newTranscoder(fmt.Sprintf("%s/%s", s.baseURL, streamName), s.resolution, s.videoBitrate, s.audioBitrate)
	}
	client := newTransmitClient(conn, s.chunkSize)
	if err := s.transcoders[streamName].addClient(client); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Printf("Adding client failed: %v", err)
	}
}
