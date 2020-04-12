package main

import (
	"fmt"
	"log"
	"os/exec"
	"sync"
)

type transcoder struct {
	url         string
	clients     []*transmitClient
	clientMutex sync.Mutex
	running     bool
	cmd         *exec.Cmd

	resolution   string
	videoBitrate string
	audioBitrate string
}

func newTranscoder(url, resolution, videoBitrate, audioBitrate string) *transcoder {
	t := new(transcoder)
	t.url = url
	t.resolution = resolution
	t.videoBitrate = videoBitrate
	t.audioBitrate = audioBitrate
	return t
}

func (t *transcoder) addClient(client *transmitClient) error {
	t.clientMutex.Lock()
	client.start()
	t.clients = append(t.clients, client)
	if !t.running {
		if err := t.start(); err != nil {
			t.clients = t.clients[:len(t.clients)-1]
			return fmt.Errorf("couldn't start transcoder: %v", err)
		}
	}
	t.clientMutex.Unlock()
	return nil
}

func (t *transcoder) start() error {
	log.Printf("Starting transcoder for %s\n", t.url)
	c := exec.Command(
		"ffmpeg", "-i", t.url,
		"-codec:v", "mpeg1video", "-s", t.resolution, "-b:v", t.videoBitrate, "-r", "30", "-bf", "0",
		"-codec:a", "mp2", "-ar", "44100", "-ac", "1", "-b:a", t.audioBitrate,
		"-f", "mpegts", "-muxdelay", "0.001",
		"-")
	pipe, err := c.StdoutPipe()
	if err != nil {
		return fmt.Errorf("starting a pipe somehow failed: %v", err)
	}
	if err := c.Start(); err != nil {
		return fmt.Errorf("running ffmpeg failed: %v", err)
	}
	t.running = true
	go func() {
		b := make([]byte, 131072)
		for t.running {
			n, err := pipe.Read(b)
			if err != nil {
				log.Printf("ffmpeg pipe read failed: %v", err)
				return
			}
			t.sendToClients(b[:n])
		}
		if err := c.Process.Kill(); err != nil {
			log.Printf("killing ffmpeg failed: %v", err)
		}
		if err := c.Wait(); err != nil {
			log.Printf("ffmpeg failed: %v", err)
		}
	}()
	return nil
}

func (t *transcoder) sendToClients(b []byte) {
	// Maybe we want to do this with fewer allocations?
	bc := append([]byte{}, b...)
	t.clientMutex.Lock()
	newClients := make([]*transmitClient, 0, len(t.clients))
	for _, c := range t.clients {
		c.send(bc) // this is fast - doesn't actually send anything
		if c.alive {
			newClients = append(newClients, c)
		}
	}
	t.clients = newClients
	if len(t.clients) == 0 {
		t.stop()
	}
	t.clientMutex.Unlock()
}

func (t *transcoder) stop() {
	log.Printf("Stopping transcoder for %s\n", t.url)
	t.running = false
}
