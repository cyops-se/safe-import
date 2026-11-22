# Security and threat model

## Assets

* The integrity and availability of the sensitive (inner) network and its systems
* The safe-import configuration (white/black lists) and audit log
* The inner CA private key (anyone holding it can impersonate any HTTPS site to the inner clients)

## Trust boundaries and assumed attackers

| Component     | Trust  | Exposed to                                    |
|---------------|--------|-----------------------------------------------|
| si-outer      | low    | Internet content (malicious servers, files)   |
| si-gatekeeper | high   | files written by si-outer                     |
| si-inner      | medium | requests from inner clients                   |
| si-engine     | high   | administrators' browsers                      |
| NATS          | high   | all components                                |

Assumed attackers:

1. **Malicious external content**: a compromised update server or a man in the middle delivers
   malware or malformed data. Mitigated by the white list, the virus scan and the separation of
   the outer/staging/inner areas.
2. **Compromised si-outer** (e.g. via a parser vulnerability): it may write arbitrary files in
   the outer area and send arbitrary messages with its own NATS identity. Mitigations:
   * NATS permissions limit what it may publish (it can not call si-gatekeeper or the inner services
     except `repos.byid`)
   * si-gatekeeper only accepts file names inside the outer area, does not follow symlinks and
     scans a private copy (no time-of-check/time-of-use window)
3. **Malicious or compromised client on the inner network**: may send arbitrary requests to
   si-inner. Only white listed URLs are fetched.
4. **Attacker with network access to the administration UI**: requires valid credentials.
   Passwords are stored as bcrypt hashes, self registration is only possible before the first
   user exists, viewers are read only and the websocket requires a valid token.

## Hardening checklist

* Run safe-import on a dedicated host, in a DMZ, behind an external firewall
* Publish the administration UI only on a management interface (`SI_ENGINE_PUBLISH`), preferably
  behind a TLS terminating reverse proxy
* Use unique, random NATS passwords (or NKeys/TLS client certificates)
* Mount the data areas `noexec,nosuid,nodev`
* Keep the CA key readable only by the si-inner user; never copy it to the inner clients
* Keep ClamAV signatures up to date (the clamav container runs freshclam)
* Review the log (failed logins, infections, `noscan` warnings, alerts) regularly

## Known limitations

These are known weaknesses that remain to be addressed:

* **Repository imports bypass the gatekeeper.** si-outer scans repositories and links them to
  the inner side itself. A compromised si-outer can therefore deliver unscanned repository
  content. This should be moved to si-gatekeeper in the same way as the HTTP flow.
* **si-outer can query si-inner** (`repos.byid`). The NATS configuration restricts it to this
  method only, but the dependency should be removed (pass the needed data in the job request).
* **si-inner connects to the real server** (port 443) to copy the certificate details for the
  generated certificate. This is an outbound connection from the inner component; it should be
  routed through si-outer.
* **White list entries with `noscan`** skip the virus scan. Every such import is logged as a
  warning together with the SHA-256 of the file.
* **The web UI uses Vue 2 / Vuetify 2**, which are end of life. A migration to Vue 3 / Vuetify 3
  is needed to keep receiving security fixes for the UI dependencies.
* **The JWT is stored in localStorage** in the browser, which makes it readable by any script
  running on the page (XSS). Moving to an HttpOnly cookie requires CSRF protection.
* The administration UI is served over plain HTTP by si-engine; use a TLS terminating proxy.

## Reporting vulnerabilities

Please report vulnerabilities privately to the maintainers (GitHub security advisories) rather than
in public issues.
