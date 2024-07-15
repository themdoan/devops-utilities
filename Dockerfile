FROM golang:1.22-buster as builder
WORKDIR /app

COPY ./ ./

RUN go mod download

RUN CGO_ENABLED=0 go build -installsuffix 'static' -o /app/devops-ulti cloudsql-monitor/main.go


FROM alpine:3
WORKDIR /app

RUN apk update; apk add git bash curl

USER nobody:nogroup

COPY --from=builder --chown=nobody:nogroup /app/devops-ulti /app
