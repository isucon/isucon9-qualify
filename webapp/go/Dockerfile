#syntax=docker/dockerfile:1

FROM golang:1.25-bookworm

RUN \
  --mount=type=cache,target=/var/lib/apt,sharing=locked \
  --mount=type=cache,target=/var/cache/apt,sharing=locked \
  apt-get update -qq && apt-get install -y \
  default-mysql-client

WORKDIR /home/webapp

COPY init.sh /home/init.sh

COPY go/go.mod /home/webapp/go.mod
COPY go/go.sum /home/webapp/go.sum
RUN go mod download -x

COPY go /home/webapp
RUN go build -o app

ENTRYPOINT [ "./app" ]
