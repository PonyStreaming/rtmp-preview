package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

type config struct {
	bind         string
	chunkSize    int
	baseURL      string
	resolution   string
	videoBitrate string
	audioBitrate string
}

func parseFlags() (config, error) {
	c := config{}
	flag.IntVar(&c.chunkSize, "chunk-size", 131072, "The size of the TS chunks to send from ffmpeg to clients")
	flag.StringVar(&c.resolution, "resolution", "350x197", "The resolution to generate streams at")
	flag.StringVar(&c.videoBitrate, "video-bitrate", "700k", "The video bitrate to stream at.")
	flag.StringVar(&c.audioBitrate, "audio-bitrate", "128k", "The video bitrate to stream at.")
	flag.StringVar(&c.bind, "bind", "0.0.0.0:8080", "The host:port to bind to.")
	flag.StringVar(&c.baseURL, "base-url", "", "The base RTMP URL, to which requests are appended.")
	flag.Parse()

	if c.baseURL == "" {
		return c, fmt.Errorf("you must specify --base-url")
	}
	return c, nil
}

func main() {
	c, err := parseFlags()
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}
	s := makeStreamer(c.baseURL, c.chunkSize, c.resolution, c.videoBitrate, c.audioBitrate)
	http.HandleFunc("/stream/", func(writer http.ResponseWriter, request *http.Request) {
		s.handle(writer, request)
	})
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	log.Printf("Serving on %s\n", c.bind)
	log.Fatal(http.ListenAndServe(c.bind, nil))
}
