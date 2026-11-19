# Upgrading

## To the hardened version (2026)

This version contains several security related changes. Read this before upgrading an
existing installation.

### Users and passwords

* Passwords are now stored as bcrypt hashes. Existing plain text passwords are converted
  automatically at the first start; the users keep their passwords.
* Users now have a role (`admin` or `viewer`). Existing users become viewers, except the oldest
  user, who is promoted to admin if there is no admin. Review the roles in the UI.
* Self registration only works while there are no users. Additional users are created by an
  administrator in the Users page.
* New passwords must be at least 10 characters.

### Login tokens

* The JWT signing key is no longer hard coded. A random key is generated in `jwt.key` (or set
  `SI_ENGINE_JWT_KEY`). All users must log in again after the upgrade.
* Tokens are valid for 8 hours by default (`SI_ENGINE_TOKEN_TIMEOUT`), instead of 30 years.

### Removed endpoints

The unauthenticated debug endpoints `/test`, `/uix`, `/msg/:text` and `/chat/:text` have been
removed. The websocket (`/ws`) now requires a valid token (`/ws?token=<jwt>`) and no longer
relays messages sent by clients.

### Gatekeeper

* Files are no longer symlinked from the outer area. They are copied to `/safe-import/staging`,
  scanned there and then moved to `/safe-import/inner`. Create the staging directory (owned by
  the gatekeeper user, mode 0750) or let si-gatekeeper create it.
* Existing symlinks in the inner area are replaced the next time the file is requested.

### Inner CA - ACTION REQUIRED

The previous CA certificate and private key (`si-inner/myCA.pem` / `myCA.key`) were committed to
the public repository. **Anyone may have a copy of that key and can impersonate any HTTPS site to
clients that trust the old CA.**

1. Generate a new CA: `install/generate-ca.sh /safe-import/ca`
2. Configure si-inner with `SI_INNER_CA_CERT` / `SI_INNER_CA_KEY`
3. Install the new `myCA.pem` on all inner clients
4. **Remove the old CA certificate from the trust store of all inner clients**

The files have been removed from the repository, but they still exist in the git history. To
remove them from the history as well (this rewrites history and requires a force push; all
clones must be re-cloned):

```sh
pip install git-filter-repo
git filter-repo --invert-paths --path si-inner/myCA.key --path si-inner/myCA.pem
git push --force --all && git push --force --tags
```

Rotating the CA (steps 1-4) is what actually removes the risk; rewriting the history only stops
further spreading.

### NATS

The components can now authenticate to NATS (`NATS_USER`/`NATS_PASSWORD`, NKeys, creds files or
TLS client certificates) and connect to another URL than localhost (`NATS_URL`). Use
[deploy/nats-server.conf](../deploy/nats-server.conf) to give each component its own user with
limited permissions.

### Build

* The SQLite driver is now pure Go; CGO / a C compiler is no longer needed.
* The built web UI is no longer committed; `si-engine/build.sh` (or the Dockerfile) builds it.
* `yarn.lock` was removed; use `npm ci` in `si-engine/ui`.
