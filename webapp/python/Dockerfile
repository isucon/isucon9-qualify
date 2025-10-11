# syntax=docker/dockerfile:1.7

FROM python:3.13-slim

WORKDIR /home/app

ENV UV_PROJECT_ENVIRONMENT=.venv

RUN --mount=type=cache,target=/var/lib/apt,sharing=locked \
    --mount=type=cache,target=/var/cache/apt,sharing=locked \
    apt-get update -qq && \
    apt-get install -y --no-install-recommends \
        build-essential \
        default-libmysqlclient-dev \
        default-mysql-client \
        pkg-config && \
    rm -rf /var/lib/apt/lists/*

RUN python -m pip install --upgrade pip && \
    python -m pip install --no-cache-dir uv

RUN mkdir -p /home/app

COPY python/pyproject.toml .

RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --compile-bytecode --no-dev

COPY init.sh /home/init.sh
RUN chmod +x /home/init.sh

COPY python/ .

ENV PATH="/home/app/.venv/bin:${PATH}"

RUN rm -f templates && ln -s /home/public templates
RUN rm -f static && ln -s /home/public static

EXPOSE 8000

ENTRYPOINT ["/home/app/.venv/bin/gunicorn", "app:app", "--bind", "0.0.0.0:8000", "--workers", "4", "--worker-class", "sync", "--access-logfile", "-", "--error-logfile", "-"]
