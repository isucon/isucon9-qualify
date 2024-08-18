export GO111MODULE=on

all: bin/benchmarker bin/benchmark-worker bin/payment bin/shipment

bin/benchmarker: cmd/bench/main.go bench/**/*.go
	go build -o bin/benchmarker cmd/bench/main.go

bin/benchmark-worker: cmd/bench-worker/main.go
	go build -o bin/benchmark-worker cmd/bench-worker/main.go

bin/payment: cmd/payment/main.go bench/server/*.go
	go build -o bin/payment cmd/payment/main.go

bin/shipment: cmd/shipment/main.go bench/server/*.go
	go build -o bin/shipment cmd/shipment/main.go

vet:
	go vet ./...

errcheck:
	errcheck ./...

staticcheck:
	staticcheck -checks="all,-ST1000" ./...

clean:
	rm -rf bin/*

init:
	$(MAKE) setup-initial-image
	$(MAKE) setup-bench-image
	$(MAKE) setup-initial-data
	$(MAKE) clean-zip

initial-data/result/initial.sql: initial-data/Dockerfile initial-data/*.tsv initial-data/*.pl
	cd initial-data && \
	docker build -t isucon9-qualify/initial-data . && \
	docker run --rm -v $(shell pwd)/initial-data/result:/opt/initial-data/result -v $(shell pwd)/initial-data/pwcache:/opt/initial-data/pwcache isucon9-qualify/initial-data && \
	docker rmi isucon9-qualify/initial-data

.PHONY: setup-initial-image
setup-initial-image:
	cd webapp/public && \
	curl -L -O https://github.com/isucon/isucon9-qualify/releases/download/v2/initial.zip && \
	unzip -qq initial.zip && \
	rm -rf upload && \
	mv v3_initial_data upload

.PHONY: setup-bench-image
setup-bench-image:
	cd initial-data && \
	curl -L -O https://github.com/isucon/isucon9-qualify/releases/download/v2/bench1.zip && \
	unzip -qq bench1.zip && \
	rm -rf images && \
	mv v3_bench1 images

.PHONY: setup-initial-data
setup-initial-data:
	cd initial-data/result && \
	curl -L -O https://github.com/isucon/isucon9-qualify/releases/download/v2/initial-data.zip && \
	unzip -qq initial-data.zip && \
	mv initial.sql ../../webapp/sql/90_initial.sql

.PHONY: clean-zip
clean-zip:
	rm -f initial-data/bench1.zip
	rm -f webapp/public/initial.zip
	rm -f initial-data/result/initial-data.zip

.PHONY: all init vet errcheck staticcheck clean
