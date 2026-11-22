# Architecture

```
          Internet                              |                Sensitive network
                                                |
   +-----------+   outer area    +--------------+--+   inner area   +-----------+
   | si-outer  | --------------> |  si-gatekeeper  | -------------> | si-inner  | <--- DNS/HTTP(S)
   | (download)|   (files)       | copy+scan+move  |   (files)      | (proxy)   |      from clients
   +-----------+                 +-----------------+                +-----------+
         |                          |         |                          |
         |                          |      clamav                        |
         |                          |                                    |
   ======+==========================+========= NATS bus ================+========
                                    |
                              +-----------+
                              | si-engine |  <--- administrators (web UI)
                              +-----------+
```

## Flow of an HTTP(S) request from the inner network

1. A client on the sensitive network resolves a name using si-inner as DNS server, and sends the
   HTTP(S) request to si-inner. HTTPS is terminated by si-inner with a certificate signed by the
   local safe-import CA (the CA certificate must be trusted by the clients).
2. si-inner matches the URL against the white, black and grey lists:
   * **black**: an alert is logged, nothing is downloaded
   * **unknown**: the URL is added to the grey list for an administrator to review
   * **white**: si-inner asks si-gatekeeper (`proxy.httpget`) for the file
3. si-gatekeeper asks si-outer (`jobs.requesturlwait`) to download the URL. si-outer stores the
   file in the **outer** area and reports the file name.
4. si-gatekeeper verifies that the reported file is inside the outer area, copies it to its
   private **staging** area (computing the SHA-256) and scans the copy with ClamAV. Infected files
   are moved to **quarantine** and reported. Clean files are moved to the **inner** area.
   Because the scanned copy is not reachable by si-outer, it can not be replaced after the scan.
5. si-inner verifies that the returned file is inside the inner area and serves it to the client.

## Data areas

| Area        | Default path               | Written by    | Read by                 |
|-------------|----------------------------|---------------|-------------------------|
| outer       | /safe-import/outer         | si-outer      | si-gatekeeper           |
| staging     | /safe-import/staging       | si-gatekeeper | si-gatekeeper, ClamAV   |
| inner       | /safe-import/inner         | si-gatekeeper | si-inner                |
| quarantine  | /safe-import/quarantine    | si-gatekeeper | administrators          |

Mount the data areas with `noexec,nosuid,nodev`.

## Message bus

All components communicate over NATS using a simple RPC convention:
`methods.<version>.<component>.<service>.<method>` with replies on `_INBOX.>`. Logs are published on
`log.<source>.<category>` and stored by si-engine. Each component has its own NATS user with
only the permissions it needs, see [deploy/nats-server.conf](../deploy/nats-server.conf).

## Repository imports

Repository mirroring (e.g. yum repositories over HTTP or SMB) is driven by si-outer, which
downloads, scans and links the result to the inner side itself. This path does **not** go
through si-gatekeeper yet, see [SECURITY.md](SECURITY.md#known-limitations).
