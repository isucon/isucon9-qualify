FROM golang:1.22-bookworm

RUN apt-get update && apt-get install -y \
  default-mysql-client \
  && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /home/webapp
WORKDIR /home/webapp

COPY init.sh /home/init.sh

COPY go/go.mod /home/webapp/go.mod
COPY go/go.sum /home/webapp/go.sum
RUN go mod download

COPY go /home/webapp
RUN go build -o app
CMD ./app
