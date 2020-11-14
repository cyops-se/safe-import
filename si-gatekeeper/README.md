# si-gatekeeper - Sanitation part of the safe-import solution
## Introduction
The gatekeeper of [safe-import](http://github.com/cyops-se/safe-import/README.md), please read more about the safe-import concept to better understand the purpose of this application.

**This is a critical application as it is tasked with the scrutiny of files downloaded from external, and potentially hostile, sources!**

It is currently highly unstable and should NOT be used in production!

The typical workflow for this application is:

- Receive download requests from si-inner
- Check the local cache if the requested resources already exists and is approved (and cache is valid)
- If the resource exist in the local cache, return with resource info (like location) to si-inner
- It the resource does not exist or the cache is invalid, forward request to si-outer
 - await response from si-outer
 - upon successful response from si-outer do
  - scan the downloaded resources for viruses (ClamAV)
  - check signatures and checksums
  - check additional meta data if available
  - if ok, make available to si-inner and send response with resource info (like location) to si-inner
