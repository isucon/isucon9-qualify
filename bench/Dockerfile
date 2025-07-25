# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.24 AS build
WORKDIR /src

RUN --mount=type=cache,target=/go/pkg/mod/ \
  --mount=type=bind,source=go.sum,target=go.sum \
  --mount=type=bind,source=go.mod,target=go.mod \
  go mod download -x

RUN --mount=type=cache,target=/go/pkg/mod/ \
  --mount=type=bind,target=. \
  CGO_ENABLED=0 go build -o /bin/benchmarker cmd/bench/main.go


FROM alpine:3.22 AS final

RUN --mount=type=cache,target=/var/cache/apk \
  apk --update add \
  ca-certificates \
  tzdata \
  && \
  update-ca-certificates

# Create a non-privileged user that the app will run under.
# See https://docs.docker.com/go/dockerfile-user-best-practices/
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

# Copy the executable from the "build" stage.
COPY --from=build /bin/benchmarker /bin/

ENTRYPOINT ["/run.sh"]
