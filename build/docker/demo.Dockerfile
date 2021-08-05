FROM golang:1.16-alpine as builder
RUN apk add build-base
COPY . /app
WORKDIR /app
RUN go env -w GOPROXY="goproxy.cn,direct"
RUN go build -ldflags="-w -s" -o demo-plugin examples/plugin-demo/main.go

FROM alpine
WORKDIR /app
COPY --from=builder /app/demo-plugin /app/demo-plugin
COPY --from=builder /app/examples/plugin-demo/html /app/html
ENTRYPOINT ["/app/demo-plugin"]
