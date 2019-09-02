all: build install-sql

build: Dockerfile *.tsv *.pl
	docker build -t isucon9-qualify/initial-data .
	docker run -v $(shell pwd)/result:/opt/initial-data/result -v $(shell pwd)/pwcache:/opt/initial-data/pwcache isucon9-qualify/initial-data

install-sql:
	install -m 0644 result/initial.sql ../webapp/sql/initial.sql
