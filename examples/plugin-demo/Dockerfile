FROM golang:1.16-alpine as builder
RUN apk add build-base
COPY . /app
WORKDIR /app
RUN go env -w GOPROXY="goproxy.cn,direct"
RUN go build -ldflags="-w -s" -o demo-plugin main.go

FROM alpine
WORKDIR /app
COPY --from=builder /app/demo-plugin /app/demo-plugin
COPY ./html ./html
COPY ./config.json ./config.json
ENTRYPOINT ["/app/demo-plugin"]
