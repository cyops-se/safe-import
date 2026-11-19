# Configuration

All components are configured with environment variables.

## All components (NATS connection)

| Variable        | Default                  | Description                                          |
|-----------------|--------------------------|------------------------------------------------------|
| `NATS_URL`      | `nats://127.0.0.1:4222`  | NATS server URL(s), comma separated                  |
| `NATS_USER`     |                          | User name                                            |
| `NATS_PASSWORD` |                          | Password                                             |
| `NATS_TOKEN`    |                          | Token authentication (alternative to user/password)  |
| `NATS_CREDS`    |                          | Path to a `.creds` file (JWT/NKey authentication)    |
| `NATS_NKEY_SEED`|                          | Path to an NKey seed file                            |
| `NATS_CA`       |                          | CA certificate for verifying the server (TLS)        |
| `NATS_CERT`, `NATS_KEY` |                  | Client certificate and key (mutual TLS)              |

The components keep trying to connect if the NATS server is not available at start-up.

## si-engine

| Variable                    | Default   | Description                                                          |
|-----------------------------|-----------|----------------------------------------------------------------------|
| `SI_ENGINE_LISTEN`          | `:7499`   | Listen address of the web UI and API                                 |
| `SI_ENGINE_DB`              | `test.db` | SQLite database file                                                 |
| `SI_ENGINE_JWT_KEY`         |           | JWT signing key (min 32 characters). Overrides the key file          |
| `SI_ENGINE_JWT_KEY_FILE`    | `jwt.key` | File with the JWT signing key. Created with a random key if missing  |
| `SI_ENGINE_TOKEN_TIMEOUT`   | `8h`      | Lifetime of a login token (Go duration, e.g. `30m`, `12h`)           |
| `SI_ENGINE_ALLOWED_ORIGINS` |           | Extra allowed websocket origins, comma separated (e.g. the dev server `http://localhost:8080`) |

Users have one of two roles:

* `admin`: may read and change everything, including users
* `viewer`: read only (may change the own password)

## si-gatekeeper

| Variable              | Default                   | Description                                         |
|-----------------------|---------------------------|-----------------------------------------------------|
| `SI_OUTER_DIR`        | `/safe-import/outer`      | Where si-outer stores downloaded files              |
| `SI_STAGING_DIR`      | `/safe-import/staging`    | Private area where files are copied and scanned     |
| `SI_INNER_DIR`        | `/safe-import/inner`      | Where vetted files are made available to si-inner   |
| `SI_QUARANTINE_DIR`   | `/safe-import/quarantine` | Where infected files are moved                      |
| `SI_CLAMDSCAN`        | `/bin/clamdscan`          | Path to clamdscan                                   |
| `SI_CLAMDSCAN_STREAM` | `0`                       | `1` streams files to clamd (clamd on another host)  |

## si-inner

| Variable                 | Default              | Description                                       |
|--------------------------|----------------------|---------------------------------------------------|
| `SI_INNER_DIR`           | `/safe-import/inner` | Files outside this directory are never served     |
| `SI_INNER_CA_CERT`       | `myCA.pem`           | CA certificate (see `install/generate-ca.sh`)     |
| `SI_INNER_CA_KEY`        | `myCA.key`           | CA private key (mode 0600)                        |
| `SI_INNER_CA_PASSPHRASE` |                      | Passphrase for a legacy encrypted PEM key         |

## si-outer

| Variable            | Default                   | Description          |
|---------------------|---------------------------|----------------------|
| `SI_CLAMDSCAN`      | `/bin/clamdscan`          | Path to clamdscan (repository scans) |
| `SI_QUARANTINE_DIR` | `/safe-import/quarantine` | Quarantine directory |
