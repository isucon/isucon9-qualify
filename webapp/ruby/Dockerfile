FROM ruby:2.7.1

# mysql client が必要なことに注意
RUN DEPS='git make gcc g++ openssl ca-certificates tzdata default-mysql-client' \
 && UTIL_DEPS='vim less wget curl' \
 && apt-get update \
 && apt-get install -y --no-install-recommends $DEPS $UTIL_DEPS

RUN cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime
RUN groupadd -r app \
 && useradd -r -g app -d /app app
RUN mkdir /app && chown app:app /app
RUN chown -R app:app /usr/local/bundle

WORKDIR /app
USER app
ENV PATH="/usr/local/bin:${BUNDLE_PATH}/bin:${PATH}"
ENV SSL_CERT_DIR=/etc/ssl/certs
ENV LANG C.UTF-8
