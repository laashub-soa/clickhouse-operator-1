FROM registry.sensetime.com/docker.io/golang:1.12.7-stretch
ARG BUILD_DIR=/go/src/github.com/mackwong/clickhouse-opertator/cmd/init-container/
COPY main.go ${BUILD_DIR}
WORKDIR ${BUILD_DIR}
RUN go build -o /usr/local/bin/maros

FROM  registry.sensetime.com/docker.io/ubuntu:16.04
COPY --from=0 /usr/local/bin/maros /usr/local/bin/maros
ENTRYPOINT ["maros"]
