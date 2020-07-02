# safe-import
https://github.com/cyops-se/safe-import

__IMPORTANT! The safe-import project, including all sibling repositories are, is in very early development and highly unstable. DO NOT USE ... yet :)__

Installation instructions can be found [here](#installation)

## Introduction
Sensitive systems, like Industrial Automation and Control Systems (IACS or ICS), must be protected from unauthorized access. This usually means isolation from other networks including the Internet.

As a paradox, it is important to move data from trusted but potentially insecure external sources.

The intent of this project is to provide mechanisms to import information to a sensitive system while mitigating the risk malware is introduced to the sensitive system, hence the name: _safe-import_.

## Principles
### Pragmatic
Root of Trust and trusted computing are simple theories that are hard to accomplish in the competitive retail world of today. Operators of IACSs seldom have possibility invest in technology and skills necessary to establish trusted computing environments in production, and this project tries to find a pragmatic balance that is possible to implement and maintain at low cost at the same time as it mitigates risks normally when it comes to import of information.

### Open Source
Without a discursion on the assurance level or lack thereof regarding open source, this project is based almost entirely on open source contributions from both well-established communities and less known repositories on github for two simple reasons;

* The sources are freely available for scrutiny, and
* It does not entail additional license costs

### Ease of use
The ambition is to provide a scripted install from a minimal, clean Linux server. The script will download and vet necessary components as far as possible in an automated environment. Once the installation completes, it is finalized and taken into operation from an (hopefully) intuitive web user interface. It has an auto-discovery feature detecting inner DNS and HTTPS requests.

#### Trust
A well-established and trusted Linux distribution is recommended, preferably CentOS. The selected Linux distribution becomes the root of trust, and standard repositories for the distribution are not vetted beyond what is already built-in the package managers. Additional packages and repositories are downloaded and vetted before installed or built.

## Concepts
### Safe import of unknown information
Acting as a termination proxy, there is no direct communicaiton between inner and outer networks. Data is pulled from external sources and all ports are closed to the outer networks.

### Outer and Inner
The basic idea is to let this solution run in a separate host that sits on the border between sensitive and potentially hostile networks. Internal filters are enabled, but an external firewall should be used to put in a DMZ with well defined communication paths.

# Architecture
## Design considerations
* outer and inner
* data only partitions
* bubblewrap
* clamav, hash/signature verifcation
* https/smb/git
* DNS

Additional Linux packages

* git
* bubblewrap
* wget
* samba

# Installation

## CentOS
```bash
git clone https://github.com/cyops-se/safe-import
cd safe-import/install
sh install_centos.sh
```

