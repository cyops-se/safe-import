# Multi-stage build of all safe-import components. Build one component with:
#   docker build --target si-engine -t safe-import/si-engine .
# or all of them with docker compose (see docker-compose.yml).

ARG GO_VERSION=1.27
ARG NODE_VERSION=22
ARG ALPINE_VERSION=3.22

# ---------------------------------------------------------------- web UI
FROM node:${NODE_VERSION}-alpine AS ui
WORKDIR /src/si-engine/ui
COPY si-engine/ui/package.json si-engine/ui/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY si-engine/ui/ ./
RUN NODE_OPTIONS=--openssl-legacy-provider npx vue-cli-service build --no-clean --dest /ui-dist

# ---------------------------------------------------------------- Go binaries
FROM golang:${GO_VERSION}-alpine AS build
RUN apk add --no-cache git
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=ui /ui-dist/ /src/si-engine/web/admin/static/
ARG GIT_COMMIT=unknown
ARG GIT_VERSION=unknown
ENV CGO_ENABLED=0
RUN for c in si-engine si-inner si-outer si-gatekeeper; do \
      go build -trimpath -ldflags "-s -w -X main.GitCommit=${GIT_COMMIT} -X main.GitVersion=${GIT_VERSION}" -o /out/$c ./$c || exit 1; \
    done

# ---------------------------------------------------------------- runtime base
FROM alpine:${ALPINE_VERSION} AS base
RUN apk add --no-cache ca-certificates tzdata \
 && addgroup -S -g 10001 si \
 && adduser -S -D -H -u 10001 -G si si \
 && mkdir -p /data /cyops/usvc/settings \
             /safe-import/outer /safe-import/inner /safe-import/staging /safe-import/quarantine \
 && chown -R si:si /data /cyops/usvc/settings /safe-import \
 && chmod 750 /safe-import/staging /safe-import/quarantine
WORKDIR /data
ENV NATS_URL=nats://nats:4222

# ---------------------------------------------------------------- components
FROM base AS si-engine
COPY --from=build /out/si-engine /usr/local/bin/si-engine
ENV SI_ENGINE_DB=/data/si-engine.db \
    SI_ENGINE_JWT_KEY_FILE=/data/jwt.key
USER si
EXPOSE 7499
ENTRYPOINT ["/usr/local/bin/si-engine"]

FROM base AS si-inner
COPY --from=build /out/si-inner /usr/local/bin/si-inner
ENV SI_INNER_DIR=/safe-import/inner \
    SI_INNER_CA_CERT=/ca/myCA.pem \
    SI_INNER_CA_KEY=/ca/myCA.key
USER si
EXPOSE 53/udp 53/tcp 80 443 8530
ENTRYPOINT ["/usr/local/bin/si-inner"]

FROM base AS si-outer
COPY --from=build /out/si-outer /usr/local/bin/si-outer
USER si
ENTRYPOINT ["/usr/local/bin/si-outer"]

FROM base AS si-gatekeeper
RUN apk add --no-cache clamav-clamdscan \
 && mkdir -p /etc/clamav \
 && printf 'TCPSocket 3310\nTCPAddr clamav\n' > /etc/clamav/clamd.conf
COPY --from=build /out/si-gatekeeper /usr/local/bin/si-gatekeeper
ENV SI_CLAMDSCAN=/usr/bin/clamdscan \
    SI_CLAMDSCAN_STREAM=1 \
    SI_OUTER_DIR=/safe-import/outer \
    SI_INNER_DIR=/safe-import/inner \
    SI_STAGING_DIR=/safe-import/staging \
    SI_QUARANTINE_DIR=/safe-import/quarantine
USER si
ENTRYPOINT ["/usr/local/bin/si-gatekeeper"]
