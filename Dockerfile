FROM golang:1.19-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN  go mod download

COPY main.go main.go
COPY internal/ internal/

RUN CGO_ENABLED=0 go build -o build/doe-bun main.go

FROM alpine:latest

COPY --from=builder /app/build/doe-bun /usr/local/bin/doe-bun

CMD doe-bun
