# 开始前先 go mod vendor
FROM golang:1.16-alpine as builder
RUN apk add build-base
COPY . /app
WORKDIR /app
RUN go env -w GOPROXY="goproxy.cn,direct"
# --tags json1 for sqlite json support
RUN go build --tags json1 -ldflags="-w -s" -o smartassistant cmd/smartassistant/main.go

FROM alpine
WORKDIR /app
COPY --from=builder /app/smartassistant /app/smartassistant
COPY ./app.yaml.example /app/app.yaml
COPY ./plugins.json.example /app/plugins.json
COPY ./static ./static
ENTRYPOINT ["/app/smartassistant"]
