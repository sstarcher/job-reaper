FROM golang:1.9-alpine AS build
RUN \
     apk update \
  && apk add git make curl \
  && (curl https://glide.sh/get | sh)
RUN mkdir -p /go/src/github.com/sstarcher/job-reaper
WORKDIR /go/src/github.com/sstarcher/job-reaper/
COPY . .
RUN \
     glide install \
  && make build

FROM golang:1.9-alpine
COPY --from=build /go/src/github.com/sstarcher/job-reaper/build/job-reaper /bin/
ENTRYPOINT ["/bin/job-reaper"]
