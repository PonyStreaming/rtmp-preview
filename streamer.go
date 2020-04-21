package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
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
	readTimeout int

	resolution   string
	videoBitrate string
	audioBitrate string
}

func makeStreamer(baseURL, resolution, videoBitrate, audioBitrate string, readTimeout int) streamer {
	return streamer{
		transcoders:  map[string]*transcoder{},
		baseURL:      baseURL,
		readTimeout:  readTimeout,
		resolution:   resolution,
		videoBitrate: videoBitrate,
		audioBitrate: audioBitrate,
	}
}

func (s *streamer) handle(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.SplitN(r.URL.Path, "/", 3)

	if len(pathParts) < 3 || pathParts[2] == "" {
		http.Error(w, "Not found", http.StatusNotFound)
	}
	streamName := pathParts[2]
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrading connection to websocket failed: %v", err)
		return
	}
	if _, ok := s.transcoders[streamName]; !ok {
		s.transcoders[streamName] = newTranscoder(fmt.Sprintf("%s/%s", s.baseURL, streamName), s.resolution, s.videoBitrate, s.audioBitrate, s.readTimeout)
	}
	client := newTransmitClient(conn)
	if err := s.transcoders[streamName].addClient(client); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Printf("Adding client failed: %v", err)
	}
}
