# safe-import - a basic Cross Domain Solution
https://github.com/cyops-se/safe-import

__IMPORTANT! The safe-import project is in early development and may exhibit unexpected and potentially unwanted behaviour. DO NOT USE IN FULL PRODUCTION... yet :)__

## Introduction
Sensitive systems, like Industrial Automation and Control Systems (IACS or ICS), must be protected physically and logically from unauthorized access, which basically means they have to be isolated from other networks and systems.

As a paradox, it is important to access data from trusted but potentially compromised external sources over the Internet, like software update repositories.

The intent of this project is to provide mechanisms to import information to a sensitive system while mitigating the risk that malware is introduced to the sensitive system, hence the name: _safe-import_. This mechanism might be called a Cross Domain Solution by some.

## Principles
### Pragmatic
Root of Trust and trusted computing are simple theories that are hard to accomplish in the competitive retail world of today. Operators of IACSs seldom have the possibility to invest in the technology and skills necessary to establish trusted computing environments in production. This project tries to find a pragmatic balance that is possible to implement and maintain at low cost, while it mitigates the risks normally associated with importing information.

### Open Source
Without a discussion on the assurance level or lack thereof regarding open source, this project is based almost entirely on open source contributions from both well-established communities and less known repositories on GitHub for two simple reasons:

* The sources are freely available for scrutiny, and
* It does not entail additional license costs

### Ease of use
The solution is installed with docker compose (or built from source) on a minimal, clean Linux server. Once running, it is configured from a web user interface. It has an auto-discovery feature detecting inner DNS and HTTP(S) requests.

### Trust
A well-established and trusted Linux distribution is recommended, e.g. Red Hat Enterprise Linux. The selected Linux distribution becomes the root of trust for this solution, and standard repositories for the distribution are not vetted beyond what is supported by the package managers. This should be an important aspect of your risk analysis when considering this solution.

## Concepts
### Safe import of unknown information
Acting as a terminating proxy, there is no direct communication between the inner and outer networks. Data is pulled from external sources, copied to an area only the gatekeeper can access, scanned for malware and only then made available to the inner side. No network ports are exposed to external networks for active exploitation.

### Outer and Inner
The basic idea is to let this solution run on a separate host that sits on the border between sensitive and potentially hostile networks. Internal filters are enabled, but an external firewall should be used to put it in a DMZ with well defined communication paths. The solution consists of a few different parts:

1. **NATS**, a light-weight and fast message bus used for communication between the components. Each component has its own NATS user with only the permissions it needs.
2. **si-outer**, executes with low privileges and pulls data from external sources on request.
3. **si-gatekeeper**, sits between the inner and outer components. It copies the content pulled by si-outer to a private staging area, scans it and moves it to the inner area.
4. **si-inner**, listens for internal DNS, HTTP and HTTPS requests and uses grey, white and black lists to determine which requests to filter out and which to forward.
5. **si-engine**, the web user interface used to manage the grey, white and black lists for DNS and HTTP(S) requests and to see the current status. If a virus is found, it is shown here.
6. **ClamAV**, the open source anti-virus used by si-gatekeeper to check the content pulled by si-outer.

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for details and [docs/SECURITY.md](docs/SECURITY.md) for the threat model.

# Installation

## Docker compose (recommended)

```sh
cp deploy/.env.example deploy/.env              # fill in the NATS passwords
sudo ./install/generate-ca.sh /safe-import/ca   # CA for HTTPS interception
sudo chown 10001 /safe-import/ca/myCA.key       # readable by the si-inner container user
docker compose --env-file deploy/.env up -d --build
```

Then open the administration UI (default http://127.0.0.1:7499/ui) and register the first user, who becomes administrator. Self registration is closed once the first user exists; further users are created by an administrator. Distribute `/safe-import/ca/myCA.pem` (never the key) to the trust store of the inner clients.

## From source

Requirements: Go (see `go.mod`), Node.js 20, a NATS server and ClamAV (`clamd` + `clamdscan`).

```sh
build/buildall.sh     # builds all components, including the web UI
go test ./...         # runs the unit tests
build/testall.sh      # starts everything locally (development only)
```

All settings are given as environment variables, see [docs/CONFIGURATION.md](docs/CONFIGURATION.md). When upgrading an existing installation, read [docs/UPGRADING.md](docs/UPGRADING.md).

# Development

* CI (GitHub Actions) builds, vets and tests the Go code, runs govulncheck, staticcheck and gosec, lints and builds the web UI and builds the Docker images.
* Dependabot keeps Go, npm, GitHub Actions and Docker dependencies up to date.
* The web UI lives in `si-engine/ui` (Vue 2 / Vuetify 2). It is built into `si-engine/web/admin/static` and embedded in the si-engine binary; the build output is not committed.
