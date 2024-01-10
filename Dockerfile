FROM golang:1.21.5 as builder

WORKDIR /app

ENV CGO_ENABLED 0
ENV GOOS linux 
ENV GOARCH amd64
#  from china use goproxy
RUN go env -w GO111MODULE=on &&\
    go env -w GOPROXY=https://goproxy.cn,direct

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY server/config.yaml ./
COPY server/main.go ./



RUN  go build -o /app/deploy-tar-server main.go

FROM busybox as runner

WORKDIR /

COPY --from=builder /app/deploy-tar-server /deploy-tar-server
COPY --from=builder /app/config.yaml /config.yaml

EXPOSE 8080

ENTRYPOINT ["/deploy-tar-server"]