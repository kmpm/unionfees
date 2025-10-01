# SPDX-FileCopyrightText: 2025 Peter Magnusosn <me@kmpm.se>
#
# SPDX-License-Identifier: CC0-1.0
FROM golang:1.25 AS builder
ARG VERSION=v0.0.0-docker
WORKDIR /app

COPY go.* ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s -X 'main.appVersion=${VERSION}'" -o ./ ./cmd/...

FROM scratch

COPY --from=builder /app/unionfees-server /unionfees-server
ENV ADDRESS=0.0.0.0:8080
EXPOSE 8080

ENTRYPOINT ["/unionfees-server"]
