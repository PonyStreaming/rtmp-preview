FROM jrottenberg/ffmpeg:4.2-scratch
COPY rtmp-preview /rtmp-preview
ENTRYPOINT ["/rtmp-preview"]
