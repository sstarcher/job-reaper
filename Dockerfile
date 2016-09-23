FROM alpine:3.4

COPY job-reaper /
ENTRYPOINT ["/job-reaper"]
