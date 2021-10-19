# 开始前先 go mod vendor
FROM golang:1.16-alpine as builder
RUN apk add build-base
COPY . /app
WORKDIR /app
ARG VERSION=latest
RUN go env -w GOPROXY="goproxy.cn,direct"
RUN go build -ldflags="-w -s -X 'main.Version=$VERSION'" -o supervisor cmd/supervisor/*.go

FROM alpine
WORKDIR /app
ARG GIT_COMMIT=unspecified
ENV GIT_COMMIT=$GIT_COMMIT
LABEL org.opencontainers.image.revision=$GIT_COMMIT
COPY --from=builder /app/supervisor /app/supervisor
ENTRYPOINT ["/app/supervisor"]
