# safe-import - a basic Cross Domain Solution
https://github.com/cyops-se/safe-import

__IMPORTANT! The safe-import project is in very early development and is likely to exhibit unexpected and potentially unwanted behaviour. DO NOT USE IN FULL PRODUCTION... yet :)__

## Introduction
Sensitive systems, like Industrial Automation and Control Systems (IACS or ICS), must be protected physically and logically from unauthorized access, which basically means it has to be isolated from other networks and systems.

As a paradox, it is important to access data from trusted but potentially compromised external sources over the Internet, like software update repositories.

The intent of this project is to provide mechanisms to import information to a sensitive system while mitigating the risk malware is introduced to the sensitive system, hence the name: _safe-import_. This mechanism might be called a Cross Domain Solution by some.

## Principles
### Pragmatic
Root of Trust and trusted computing are simple theories that are hard to accomplish in the competitive retail world of today. Operators of IACSs seldom have possibility invest in technology and skills necessary to establish trusted computing environments in production, and this project tries to find a pragmatic balance that is possible to implement and maintain at low cost at the same time as it mitigates risks normally when it comes to import of information.

### Open Source
Without a discursion on the assurance level or lack thereof regarding open source, this project is based almost entirely on open source contributions from both well-established communities and less known repositories on github for two simple reasons;

* The sources are freely available for scrutiny, and
* It does not entail additional license costs

### Ease of use
The ambition is to provide a scripted install from a minimal, clean Linux server. The script will download and vet necessary components as far as possible in an automated environment. Once the installation completes, it is finalized and taken into operation from an (hopefully) intuitive web user interface. It has an auto-discovery feature detecting inner DNS and HTTPS requests.

### Trust
A well-established and trusted Linux distribution is recommended e.g. Red Hat Enterprise Linux. The selected Linux distribution becomes the root of trust for this solution and standard repositories for the distribution are not vetted beyond what is supported by the package managers. Additional packages and repositories are downloaded and vetted before installed or built. This should be an important aspect of your risk analysis when considering this solution.

## Concepts
### Safe import of unknown information
Acting as a termination proxy, there is no direct communicaiton between inner and outer networks. Data is pulled from external sources and stored on a non-executable partition. No network ports are exposed to external networks for active exploitation.

### Outer and Inner
The basic idea is to let this solution run in a separate host that sits on the border between sensitive and potentially hostile networks. Internal filters are enabled, but an external firewall should be used to put in a DMZ with well defined communication paths. The solution consist of a few different parts;

1. NATS, a light-weight and fast message queue which is used to communicate between the different components
2. si-outer, a component that executes in a low-privileged user space and is responsible to pull data from external sources on request.
3. si-gateway, sits between the inner and outer components and is tasked with vetting the content pulled by si-outer before making it available to si-inner.
4. si-inner, listens for internal DNS, HTTP and HTTPS requests and uses a grey, white, black list mechanism to determine which requests to filter out and which to forward. This process executes in a privileged user space at it listens to port 53
5. si-engine, a user interface used to manage the grey, white and black lists for DNS and HTTP(S) requests, and to see the current status of things. If a virus is found, it will be shown in this user interface.
6. ClamAV, an open source anti-virus used by si-gateway to check the content pulled by si-outer

# Installation
Simple scripts for building and testing are provided in the build folder, but work is in progress to provide docker images for this solution to make it easier to use.
