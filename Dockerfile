FROM registry.sensetime.com/docker.io/golang:1.12.7-stretch
ARG BUILD_DIR=/go/src/github.com/mackwong/clickhouse-operator/
COPY . ${BUILD_DIR}
WORKDIR ${BUILD_DIR}
RUN go build -o ${BUILD_DIR}/bin/clickhouse-all-in-one ./cmd/app/app.go

FROM  registry.sensetime.com/docker.io/ubuntu:16.04
ARG BUILD_DIR=/go/src/github.com/mackwong/clickhouse-operator/
COPY --from=0 ${BUILD_DIR}/bin/clickhouse-all-in-one /usr/local/bin/clickhouse-all-in-one
COPY clickhouse.yaml /etc/broker/clickhouse.yaml
ENTRYPOINT ["clickhouse-all-in-one"]
