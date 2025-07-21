FROM python:3.13-slim

WORKDIR /home/app

RUN apt-get update && apt-get install -y \
    gcc \
    default-libmysqlclient-dev \
    default-mysql-client \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

COPY init.sh /home/init.sh
RUN chmod +x /home/init.sh

COPY python/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY python/ .

RUN rm -f templates && ln -s /home/public templates
RUN rm -f static && ln -s /home/public static

EXPOSE 8000

CMD ["gunicorn", "--bind", "0.0.0.0:8000", "--workers", "4", "--worker-class", "sync", "--access-logfile", "-", "--error-logfile", "-", "app:app"]
