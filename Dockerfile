FROM golang:alpine AS builder

WORKDIR /src
COPY . /src

RUN go env -w GOPROXY=https://goproxy.cn,direct &&  \
    go build -o miwifi_exporter main.go

FROM alpine

WORKDIR /app
COPY --from=builder /src/miwifi_exporter /app/miwifi_exporter
COPY --from=builder /src/config.json /app/config.json
EXPOSE 9001

CMD ["/app/miwifi_exporter"]