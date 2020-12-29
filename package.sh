#!/bin/env bash
set -e

CHARTSREPO="https://registry.sensetime.com/chartrepo/diamond-dev/charts/"

function build_and_push_image() {
    make push
}

function push_charts() {
    cd install/helm
    if [ -z ${CI_COMMIT_TAG} ];then
      sed -i "s/^version:.*/&-${CI_COMMIT_SHORT_SHA}/g" clickhouse-broker/Chart.yaml
      sed -i "s/^version:.*/&-${CI_COMMIT_SHORT_SHA}/g" clickhouse-operator/Chart.yaml
    fi

    helm repo add --username $USER --password "$PASSWORD" diamond-dev https://registry.sensetime.com/chartrepo/diamond-dev
    helm push clickhouse-broker diamond-dev --username $USER --password "$PASSWORD" && \
    helm push clickhouse-operator diamond-dev --username $USER --password "$PASSWORD"

    cd -
}

function release() {
    version=$(grep version install/helm/clickhouse-operator/Chart.yaml | awk '{print $2}')
    operator_url=${CHARTSREPO}/clickhouse-operator-${version}.tgz
    version=$(grep version install/helm/clickhouse-broker/Chart.yaml | awk '{print $2}')
    broker_url=${CHARTSREPO}/clickhouse-broker-${version}.tgz
    auto_release --id ${CI_PROJECT_ID} --tag ${CI_COMMIT_TAG} --operator ${operator_url} --broker ${broker_url} --service ${CI_PROJECT_NAME}
}

function set_vars() {
     if [ ! -z ${CI_COMMIT_TAG} ];then
      TAG=`git tag --sort=committerdate | tail -n 1`
      IMAGE="registry.sensetime.com/diamond/service-providers/clickhouse-all-in-one"
    else
      TAG="$CI_COMMIT_BRANCH"-"$CI_COMMIT_SHORT_SHA"
      IMAGE="registry.sensetime.com/diamond-dev/service-providers/clickhouse-all-in-one"
      export IMAGE=$IMAGE
    fi

    mv $DOCKER_CONFIG config.json

    export TAG=$TAG
    export IMAGE=$IMAGE
    export DOCKER_DIR=$(pwd)
}

function print_info() {
    echo -e "\n\n\n"
    echo "clickhouse operator && broker has been uploaded:"
    echo -n "Operator: "
    grep version install/helm/clickhouse-operator/Chart.yaml
    echo -n "Broker: "
    grep version install/helm/clickhouse-broker/Chart.yaml
    echo -e "\n\n\n"
}


set_vars
build_and_push_image
push_charts
release
print_info



