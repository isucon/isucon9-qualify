FROM golang:1.22

RUN mkdir -p /opt/go
WORKDIR /opt/go

COPY initial-data /initial-data
COPY webapp/public/static /static

COPY go.mod /opt/go/go.mod
COPY go.sum /opt/go/go.sum
RUN go mod download

COPY cmd/ /opt/go/cmd
COPY bench/ /opt/go/bench

RUN go build -o benchmarker cmd/bench/main.go

ENTRYPOINT ["/opt/go/bench/run.sh"]
