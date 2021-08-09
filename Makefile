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

PROJECT_NAME=isucon

# ssh用の公開鍵をセットする
set-ssh:
	@curl https://github.com/Tatsuemon.keys >> ~/.ssh/authorized_keys
	@xcurl https://github.com/tarao1006.keys >> ~/.ssh/authorized_keys
	@curl https://github.com/yuyafukuchi.keys >> ~/.ssh/authorized_keys

# git configとssh 設定
git-config:
	# githubの設定
	@git config --global user.name "Tatsuemon"
	@git config --global user.email "i10mann-110@ezweb.ne.jp"

	# ssh Hostの設定
	@echo "
	Host github github.com
	HostName github.com
	PreferredAuthentications publickey
	IdentityFile ~/.ssh/id_git_rsa
	User git
	" >> ~/.ssh/config

	# ssh-key
	@ssh-keygen -t rsa -N "" -f ~/.ssh/id_git_rsa
	@echo ~/.ssh/id_git_rsa.pub

# /etc/nginx, /etc/mysqlを持ってくる, 基本的に一回だけ
mv:
	@if [ ! -d /etc/nginx1 ]; then\
		sudo cp -r /etc/nginx/ ~/${PROJECT_NAME}/; \
		sudo cp -r /etc/nginx/ /etc/nginx1/; \
		sudo cp -r /etc/mysql ~/${PROJECT_NAME}/; \
		sudo cp -r /etc/mysql/ /etc/mysql1/; \
	fi
	

# github管理下にあるものをcpしてreload
reload-nginx:
	sudo cp -r ~/${PROJECT_NAME}/nginx/ /etc/
	sudo nginx -t
	sudo nginx -s reload

# github管理下にあるものをcpしてreload
reload-mysql:
	sudo cp -r ~/${PROJECT_NAME}/mysql /etc/
	sudo systemctl restart mysql.service


# 計測で使用するもの
install-tools:
	# alp
	@wget https://github.com/tkuchiki/alp/releases/download/v1.0.7/alp_linux_amd64.zip && \
	unzip alp_linux_amd64.zip && \
	sudo mv alp /usr/local/bin/ && \
	rm -rf alp_linux_amd64.zip

	# pt-query-digest
	@sudo apt-get install -y gnupg2 && \
		wget https://repo.percona.com/apt/percona-release_latest.$(shell lsb_release -sc)_all.deb
		sudo dpkg -i percona-release_latest.$(shell lsb_release -sc)_all.deb

	@sudo apt-get update && \
		sudo apt-get install -y percona-toolkit && \
		rm -rf  percona-release_latest.$(shell lsb_release -sc)_all.deb
	pt-query-digest --version
