# syntax=docker/dockerfile:1

FROM golang:1.22-bookworm
WORKDIR /src

RUN --mount=type=cache,target=/go/pkg/mod/ \
  --mount=type=bind,source=go.sum,target=go.sum \
  --mount=type=bind,source=go.mod,target=go.mod \
  go mod download -x

RUN --mount=type=cache,target=/go/pkg/mod/ \
  --mount=type=bind,target=. \
  CGO_ENABLED=0 go build -o /bin/benchmarker cmd/bench/main.go

ARG UID=10001
RUN adduser \
  --disabled-password \
  --gecos "" \
  --home "/nonexistent" \
  --shell "/sbin/nologin" \
  --no-create-home \
  --uid "${UID}" \
  appuser
USER appuser

COPY initial-data /initial-data
COPY webapp/public/static /static

COPY bench/run.sh /run.sh

ENTRYPOINT ["/run.sh"]
