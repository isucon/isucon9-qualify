# syntax=docker/dockerfile:1

FROM node:22-bookworm AS builder

WORKDIR /home/webapp

# Copy package files
COPY nodejs/package.json nodejs/package-lock.json* /home/webapp/
RUN npm ci

# Copy source files and build
COPY nodejs /home/webapp
RUN npm run build

# Production stage
FROM node:22-bookworm-slim

RUN \
  --mount=type=cache,target=/var/lib/apt,sharing=locked \
  --mount=type=cache,target=/var/cache/apt,sharing=locked \
  apt-get update -qq && apt-get install -y \
  default-mysql-client \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /home/webapp

# Copy init script
COPY init.sh /home/init.sh

# Copy only production dependencies
COPY nodejs/package.json nodejs/package-lock.json* /home/webapp/
RUN npm ci --omit=dev

# Copy built application and public files
COPY --from=builder /home/webapp/dist /home/webapp/dist
COPY public /home/webapp/public

EXPOSE 8000

CMD ["npm", "start"]
