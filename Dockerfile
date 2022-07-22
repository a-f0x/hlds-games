FROM golang:1.18-alpine3.15 AS build-env

WORKDIR /build
COPY . .

RUN cd /build/cmd/GAME_AGENT && \
    CGO_ENABLED=0 go build -v -o game_agent

FROM scratch AS export-build
COPY --from=build-env /build/cmd/GAME_AGENT/game_agent /