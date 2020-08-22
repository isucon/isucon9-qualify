export GO111MODULE=on
export GOOGLE_CLOUD_PROJECT=isucon-gigawatts
export ENABLE_TRACE=true
export ENABLE_PROFILE=true
export GOOGLE_APPLICATION_CREDENTIALS=/home/${USER}/serviceaccount-key.json
export CAMPAIGN=0

.PHONY: build
build: *.go
	@echo "### ENVIOROMENT INFO ###"
	@echo "CAMPAIGN = ${CAMPAIGN}\nENABLE_TRACE = ${ENABLE_TRACE}\nENABLE_PROFILE = ${ENABLE_PROFILE}"
	@go build -o isucari

.PHONY: run
run: build
	@./isucari

.PHONY: vendor
vendor:
	go mod vendor
