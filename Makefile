export GO111MODULE=on

all: bin/benchmarker bin/payment bin/shipment

bin/benchmarker: cmd/bench/main.go bench/**/*.go external/payment/*.go
	go build -o bin/benchmarker cmd/bench/main.go

bin/payment: cmd/payment/main.go external/payment/*.go bench/server/*.go
	go build -o bin/payment cmd/payment/main.go

bin/shipment: cmd/shipment/main.go bench/server/*.go
	go build -o bin/shipment cmd/shipment/main.go

vet:
	go vet ./...

errcheck:
	errcheck ./...

staticcheck:
	staticcheck -checks="all,-ST1000" ./...
