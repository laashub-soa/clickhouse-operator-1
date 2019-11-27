FROM registry.sensetime.com/docker.io/golang:1.12.7-stretch
ARG BUILD_DIR=/go/src/github.com/mackwong/clickhouse-operator/
COPY . ${BUILD_DIR}
WORKDIR ${BUILD_DIR}
RUN go build -o ${BUILD_DIR}/bin/broker ./cmd/clickhouse-broker/main.go

FROM  registry.sensetime.com/docker.io/ubuntu:16.04
ARG BUILD_DIR=/go/src/github.com/mackwong/clickhouse-operator/
COPY --from=0 ${BUILD_DIR}/bin/broker /usr/local/bin/broker
COPY clickhouse.yaml /etc/broker/clickhouse.yaml
ENTRYPOINT ["broker"]
