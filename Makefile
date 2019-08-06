export GO111MODULE=on

all: bin/benchmarker bin/payment bin/shipment

bin/benchmarker: bench/*/*.go
	go build -o bin/benchmarker cmd/bench/main.go

bin/payment: external/payment/*.go
	go build -o bin/payment cmd/payment/main.go

bin/shipment: external/shipment/*.go
	go build -o bin/shipment cmd/shipment/main.go
