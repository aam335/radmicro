FROM golang:alpine AS builder

RUN apk update
RUN apk add git libc-dev gcc

WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go mod tidy
RUN go build -o main .

WORKDIR /dist
RUN cp /build/main .
WORKDIR /data
ADD config.toml .
########################################################
FROM alpine:latest

COPY --chown=0:0 --from=builder /dist /
COPY --chown=65534:0 --from=builder /data /data

USER 65534
WORKDIR /data

EXPOSE 1812/udp

ENTRYPOINT ["/main"]

