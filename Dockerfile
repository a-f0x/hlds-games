FROM golang:1.18-alpine3.15 AS build-env

WORKDIR /build
COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./pkg ./pkg
COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum

RUN cd /build/cmd/GAME_AGENT && \
    CGO_ENABLED=0 go build -v -o game_agent && \
    cd /build/cmd/GAME_MONITOR && \
    CGO_ENABLED=0 go build -v -o monitoring

FROM scratch AS export-build
COPY --from=build-env /build/cmd/GAME_AGENT/game_agent /
COPY --from=build-env /build/cmd/GAME_MONITOR/monitoring /