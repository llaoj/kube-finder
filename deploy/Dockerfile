# FROM registry.cn-beijing.aliyuncs.com/llaoj/golang:1.18
FROM golang:1.18

WORKDIR /usr/src/app

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn

COPY . .
COPY deploy/config.prod.yaml /etc/kube-finder/config.yaml
RUN go build -v -o /usr/bin/ ./...

CMD ["kube-finder"]