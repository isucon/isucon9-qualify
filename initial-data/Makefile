all:
	docker build -t isucon9-qualify/initial-data .
	docker run -v $(shell pwd)/result:/opt/initial-data/result isucon9-qualify/initial-data
