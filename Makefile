.PHONY: lint test coverage build image push deploy-operator deploy-broker install release package clean uninstall all-clean

IMAGE ?= registry.sensetime.com/diamond/service-providers/clickhouse-operator
TAG ?= latest
PULL ?= Always

lint: ## Run all the linters
	golangci-lint run --fast --deadline 3m  --skip-dirs vendor ./...

test:
	echo 'mode: atomic' > coverage.txt && go test -covermode=atomic -coverprofile=coverage.txt -v -run="Test*" -timeout=30s ./...

build: clean
	go build -o bin/broker -ldflags "-X main.Version=$(shell git describe)" cmd/manager/main.go

image:
	docker build --no-cache . -t "$(IMAGE):$(TAG)"

coverage: test
	go tool cover -html=coverage.txt -o coverage.html

clean: ## Cleans up build artifacts
	rm -rf bin/*

uninstall: ## Uninstall operator and broker
	helm delete --purge clickhouse-service-broker
	helm delete --purge clickhouse-operator

all-clean: uninstall## Delete all binary and resources related to clickhouse service
	# WARNING!!! This will also delete your clusters which use our CRD
	rm -rf bin/*
	kubectl delete crd clickhouse.sensetime.com

push: image ## Pushes the image to docker registry
	docker push "$(IMAGE):$(TAG)"

deploy-operator: ## Deploys operator with helm
	helm upgrade --install clickhouse-operator helm/clickhouse-operator --namespace clickhouse-system

generate: ## Deploys operator with helm
	 operator-sdk generate k8s
	 operator-sdk generate openapi

deploy-broker: ## Deploys broker with helm
	helm upgrade --install clickhouse-service-broker \
	helm/clickhouse-service-broker --namespace clickhouse-system \
	--set image="$(IMAGE):$(TAG)",imagePullPolicy="$(PULL)"

install: deploy-operator deploy-broker ## install components of clickhouse serivces
	# Installation: clickhouse operator installed with Helm.
	# Installation: clickhouse broker installed with Helm.
