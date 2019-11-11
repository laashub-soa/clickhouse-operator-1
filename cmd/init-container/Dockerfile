FROM registry.sensetime.com/docker.io/golang:1.12.7-stretch
ARG BUILD_DIR=/go/src/gitlab.bj.sensetime.com/diamond/service-providers/cassandra
COPY . ${BUILD_DIR}
WORKDIR ${BUILD_DIR}
RUN make build

FROM  registry.sensetime.com/docker.io/ubuntu:16.04
ARG BUILD_DIR=/go/src/gitlab.bj.sensetime.com/diamond/service-providers/cassandra
COPY --from=0 ${BUILD_DIR}/bin/* /usr/local/bin/
