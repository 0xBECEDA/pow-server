FROM golang:1.23 AS builder

RUN mkdir build
COPY . /build

WORKDIR /build
RUN go mod download

WORKDIR /build/cmd/client
RUN  go build -o main

WORKDIR /build/cmd/server
RUN go build  -o main

FROM golang:1.23

COPY --from=builder /build/cmd/client/main /client
COPY --from=builder /build/cmd/server/main /server