# si-inner - Part of the safe-import solution
## Introduction
The inner interface of [safe-import](http://github.com/cyops-se/safe-import/README.md), please read more about the safe-import concept to better understand the purpose of this application.

It exposes DNS and common file retrieval services used to capture inner requests and notify the outer side of wanted resources. Inner requests not previously processed will get a 'resource missing' kind of a response to encourage it to try again later. Once it receives an approval from the outer side, it will serve the resources to the one requesting it.

It is not uncommon in IACSs to have a local Microsoft Active Directory for central user account and Windows host management. In such configurations, the Domain Controllers hosting the Active Directory also acts as local DNS servers for the hosts in the Windows domain. All DNS requests are targeting names outside of the local Windows domains will forwarded by default to external DNS servers. As this is an unwanted behavior in an IACS, forwarding should be disabled or at least filtered by an internal or perimeter filter (firewall).

That is of course unless you intend to introduce safe-import. With safe-import, you configure the local DNS to forward requests to the safe-import host instead, which will either reply with its own IP address if the request is approved, or with NXDOMAIN if the domain name has been black- or whitelisted (white block).

The application making the DNS request will then make the next request to the safe-import host and si-inner can pick it up if it is one of the supported services. Approved requests are forwarded to the outer parts of safe-import for actual downloading and sanitation before the requested content can be provided to the client requesting it.

### Example - Approved single file request
In this example, the domain name and URL has already been classifed as approved white, meaning they will be processed by safe-import

- Client within IACS is configured to automatically update some software from the URI https://x.y.z/download/v1/somesoftware.tar.gz
- Client makes a DNS request to resolve x.y.z to the Domain Controller DNS
- The Domain Controller DNS forwards the request to safe-import host
- si-inner receives the request, determines if the domain is approved or not. In this case it is approved, and si-inner responds with the IP of safe-import
- Client make a HTTPS request for the URI https://x.y.z/download/v1/somesoftware.tar.gz. This request will be directed to the safe-import host as the x.y.z name resolved to safe-import
- si-inner receives the HTTPS request, finds the URL to be approved and requests the resource from the outer part of safe-import
- The outer part downloads and sanitizes the file before making it available for si-inner
- si-inner sends the content as reply to the Client


## White, Black and Grey
The number of requests can be overwhelming, which makes it important to use a systematic approach to handling them, which is where the use of white, black and grey lists are useful.

### Grey
Requests not already classified are put in the grey list and clients receive a 404 or NXDOMAIN. A human operator must then decide if the request is white or black. It is important to keep the grey list clear as it eventually becomes a strong indicator of compromise due to the static nature of an IACS.

### Black
Requests in the black list are known bads and should be alerted upon as it is a definite indication of compromise or misconfiguration. From a compliance perspective, it can for example indicate violation of configuration change management policies.

### White
Anything that is determined to be non-hostile (although potentially annoying) should be put in the white list. Most white entries are however accepted but blocked, meaning it is a non-hostile request which should be ignored. One example of such requests can be XBOX update requests from a Windows host. White entries also includes those that should be handled through the sandbox.