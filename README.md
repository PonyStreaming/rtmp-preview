# rtmp-preview

`rtmp-preview` is a program that uses ffmpeg to read an RTMP stream,
transcode the stream to a tiny, low bitrate MPEG-TS (MPEG1 + MP2) stream,
and stuffs the result into a websocket. The latency incurred through this
process is negligible (well under a second).

The result is intended for use with [jsmpeg](https://jsmpeg.com/), a
JavaScript-based MPEG1/MP2 decoder.
