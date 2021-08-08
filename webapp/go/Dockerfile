FROM golang:1.12.7
WORKDIR /app

RUN apt update \
    apt install -y mariadb-client

COPY ./go.mod ./
RUN go mod download

CMD /bin/sh -c "while sleep 1000; do :; done"