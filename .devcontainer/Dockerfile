FROM ubuntu:18.04

RUN apt update && apt upgrade -y

RUN apt install -y python3 \
    python3-pip \
    python3-dev \
    libffi \
    mysql-server \
    mysql-client \
    default-libmysqlclient-dev \
    git \
    wget \
    curl \
    build-essential


COPY requirements.txt .
RUN pip3 install -r requirements.txt

# Go 1.13 をインストール (M1 mac の場合は AMD を ARM に変更)
RUN wget https://dl.google.com/go/go1.16.5.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.16.5.linux-amd64.tar.gz

RUN echo 'export PATH=$PATH:/usr/local/go/bin' >> /root/.bashrc

