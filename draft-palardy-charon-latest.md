---
title: "CHARON - Cached Host Address Resolution for Outside Networks"
abbrev: "CHARON"
category: info

docname: draft-palardy-charon-latest
submissiontype: IETF  # also: "independent", "editorial", "IAB", or "IRTF"
number:
date:
consensus: true
v: 3
area: AREA
workgroup: WG Working Group
keyword:
 - next generation
 - unicorn
 - sparkling distributed ledger
venue:
  group: WG
  type: Working Group
  mail: WG@example.com
  arch: https://example.com/WG
  github: USER/REPO
  latest: https://example.com/LATEST

author:
 -
    fullname: Andrew Palardy
#    organization: Independent
    email: andrew@apalrd.net

normative:

informative:

...

--- abstract

Cached Host Address Resolution for Outside Networks is a method for combining DNS-based address translation and SIIT-based packet translation to allow legacy IPv4 hosts to access IPv6-only outside networks. This permits legacy IPv4-only devices to continue to function even as the global internet transitions to providing services over only IPv6.

To fully transition Internet routing from IPv4 to IPv6, translation is currently used to permit hosts which speak only IPv4 or IPv6 to speak to hosts which speak only the other network protoocol. Current translation methods place the burden of IPv4-IPv6 translation on the IPv6-speaking network, thus requiring all inter-network traffic to support IPv4 as a global fallback. It is desirable in the long term to shift this 'default' to requiring all inter-area traffic to support IPv6, with IPv4 support gradually declining. However, it is forseeable that legacy IPv4-only devices will continue to operate for some time and will still require translation by their own network operators.

CHARON provides this method, allowing a network operator with IPv4-only clients to provide access to IPv6-only services by dynamically mapping IPv6-only services resolved via DNS to a local-use IPv4 address on a per address basis. It additionally shifts the burden for providing translation services onto the network supporting legacy IPv6 devices. 

--- middle

# Conventions and Definitions

{::boilerplate bcp14-tagged}

# Introduction

Currently, all hosts on the Internet implement either IP version 4, IP version 6, or both (a "dual stack" configuration). As IPv4 address resources have been exhausted for some time, IPv4 re-use is becoming a larger concern. While IPv6 is the preferred option, IPv6 does not free IPv4 resources - as the Internet operates both versions in parallel, it is currently expected that all services will be accessible via IPv4, and therefore all clients must be able to utilize IPv4, continuing the demand for global IPv4 resources. It is not currently expected that all services will be accessible via IPv6, so networks are able to continue operating using only IPv4.

As operating both protocol families is more difficult than operating only one, it is desirable to implement only the IPv6 protocol. However, networks which operate only the IPv6 protocol still must inter-operate with networks which only operate the IPv4 protocol, requiring protocol translation. Currently, the assumption is that traffic over the internet may always fall back to IPv4, therefore, the burden is on IPv6-only network operators to provide translation services on behalf of IPv4-only networks they are interacting with.

The goal of CHARON is to flip this assumption, and allow IPv4-only networks to provide translation services on their own, freeing IPv6-only networks to remain IPv6-only. 

## Requirement for IP/ICMP Translation

If a source wants to send a packet to dest:
  \      source
   \   v4 | ds| v6
d   \______________
e v4|  B  | B |  C
s ds|  B  | A |  A
t v6|  D  | A |  A

There are four possible scenarios which could be encountered:
A. Both support IPv6 -> IPv6 is preferred and IPv6 used
B. Both support IPv4 -> IPv4 is the global fallback and IPv4 is used
C. Src only supports v6 and dest only supports v4 -> Translation required
D. Src only supports v4 and dest only supports v6 -> Translation required

This translation requirement leaves two options for where to tackle cases C and D using protocol translation - translator provided by the source network or by the destination network.

## Existing Architectures
Existing RFCs define architectures for several translaton scenarios, focused on permitting IPv6-only hosts to interact with a dual-stack Internet, and with translators operated by the IPv6-only network.

For networks with v6-only clients establishing v4 connections, 464XLAT (RFC6877) allows v6-only sources which support a client-side translator (CLAT) to speak to v4-only destinations. This works cleanly across any higher layer protocol and does not depend on DNS or any other side-channel protocol, and is widely deployed by client ISPs. Without a client-side translator, DNS64 (RFC6147) may be used to direct clients to the NAT64 function via DNS. 

For networks with v6-only servers accepting v4 connections, SIIT (RFC7915) to allow the v4-internet to speak to v6-only datacenters, and this is widely deployed as SIIT-DC (RFC7755)

For networks with v4-only servers accepting v6 connections, SIIT (RFC7915) can technically work, but there is no standard for how to compress the IPv6 source address into an IPv4 field to send to the server. TAYGA (github.com/apalrd/tayga) solves this problem by dynamically allocating IPv4 addresses out of a pool, for IPv6 clients. Application-layer gateways (i.e. HTTP proxies) may also be used here. 

## Cached Host Address Resolution for Outside Networks
As the IPv4 address space is insufficient for full IPv6 translation, some mechanism must be used to discover which IPv6 addresses are required and translate only those addresses. This translation 'window' is accessible via an IPv4 address pool, which exists within the IPv4 address space. This IPv4 address pool is routed to the translator. The translator creates mapping entires for single hosts (IPv4 <-> IPv6 pair) on demand within this address pool.

        +---------------------------+
        |        IPv4 Network       |
        |                           |
        |  +-------+   +-------+    |
        |  |Host A |   |Host B |    |
        |  +---+---+   +---+---+    |
        |      |           |        |
        |      +-----+-----+        |
        |            |              |
        +------------+--------------+
                     |
          +----------+----------+
          |                     |
 (DNS A)  |                     |  (IPv4 Traffic)
          v                     v
      +---+---------------------+---+
      |     IPv4/IPv6 Translator    |
      +---+---------------------+---+
          ^                     ^
(DNS AAAA)|                     |  (IPv6 Traffic)
          |                     |
          +----------+----------+
                     |
        +------------+--------------+
        |        IPv6 Internet      |
        |                           |
        |    +---------------+      |
        |    | IPv6 Servers  |      |
        |    +---------------+      |
        |                           |
        +---------------------------+

To allow the IPv4 host to initiate a connection to an IPv6 host, the IPv4 host must first perform a DNS A request. In response to the DNS request, the translator will perform a corresponding AAAA request, and create new mapping entries in response to DNS requests by IPv4 hosts, returning the new mapping entries to the IPv4 host in the DNS A response. This permits full network layer translation, without any higher layer protocol gateways. 

To allow an IPv6 host to initiate a connection to an IPv4 host, the translator creates new mapping entires in response to the first IPv6 packet seen by that host, allowing the mapping entry to be used as a source address in the IPv4 packet. This direction does not depend on DNS.

Mapping entries are retained for subsequent connections, and are re-used by any traffic using the same IPv6 address. They may time out after some period of inactivity. 

# Address Mapping Table
This RFC extends the address mapping methods used by Stateless IP/ICMP Translation (SIIT) [RFC7915] to add a Dynamic Address Mapping option. 

With this option, a SIIT translator may translate addresses using any of the following methods:
- Encode the IPv4 address into an IPv6 address [RFC6052]
- Explicitly-configured IPv4 <-> IPv6 address mapping [RFC7757]
- Dynamically-configured IPv4 <-> IPv6 address mapping [This Document]

All three methods may be used simultaneously, see [Deployment Architectures] for examples. 

Dynamic mapping entries map a single IPV4 and a single IPv6 (prefix length 32 and 128 respectively), and MUST be re-used for all cases where that IPv6 host address is encountered (including IPv6-initiated mappings and multple DNS hostnames)

## IPv6 initiated Dynamic Mapping

IPv6-initiated dynamic mappings are created when an IPv6 packet is received by the translator such that the destination IPv6 address can be translated using an existing mapping method (encoded, explicit, or dynamic), but the source IPv6 address does not match any existing mapping entry.

In this case, the translator MUST allocate an available IPv4 address from its dynamic pool and create a new dynamic mapping entry binding the source IPv6 address to the allocated IPv4 address. This mapping is then used to translate the source address for all subsequent packets in this flow and any future traffic involving the same IPv6 address.

If no IPv4 addresses are available in the dynamic pool, the translator MUST drop the packet and SHOULD generate an appropriate ICMP error indicating address translation failure.

## IPv4+DNS initiated Dynamic Mapping

IPv4-initiated dynamic mappings are created in response to DNS queries from IPv4-only hosts. When an IPv4 host issues a DNS A query, the translator MUST perform corresponding upstream DNS queries for the corresponding AAAA record.

The translator MAY choose to perform a second query for A records, and return that response if it exists, according to local policy.

If one or more AAAA records are returned, the translator MUST, for each IPv6 address in the AAAA response, check whether a corresponding mapping already exists in either the explicit or dynamic mapping tables. If no mapping exists for a given IPv6 address, the translator MUST allocate an IPv4 address from the dynamic pool and create a new mapping entry binding that IPv6 address to the allocated IPv4 address.

The translator MUST then synthesize A records using the mapped IPv4 addresses corresponding to the AAAA records and return them to the IPv4 host. If multiple AAAA records are present, each MUST be mapped independently.

The TTL of the returned record MUST be the lesser of the TTL of the AAAA record used for synthesis, and the expiration time of the dynamic mapping entry.

## Reverse DNS

The translator MUST respond to reverse DNS (PTR) queries within the address space of its dynamic IPv4 pool. Upon receiving a query for an address within this pool (in-addr.arpa), the translator MUST look up the corresponding IPv6 address using its mapping tables.

If a mapping exists, the translator MUST perform a reverse DNS query (in6.arpa) for the associated IPv6 address and return the resulting PTR record(s) to the requester. If multiple PTR records are returned, all applicable records SHOULD be included in the response.

If no mapping exists for the queried IPv4 address, the translator SHOULD return an appropriate negative DNS response (e.g., NXDOMAIN).

## Mapping Timeouts

Each dynamic mapping entry MUST track activity timestamps, including the most recent packet observed in either translation direction and the most recent DNS query that resulted in the associated IPv6 address.

A mapping entry MUST be retained and reused as long as it remains active. Activity is defined as:

Any IPv4 or IPv6 packet translated using the mapping, in either direction.
Any DNS query (A or AAAA) that results in the associated IPv6 address.

The translator MUST retire (delete) a mapping entry after a configurable period of inactivity. This timeout value SHOULD be configurable by the operator and SHOULD balance efficient reuse of the IPv4 pool with stability of address mappings.

When a mapping is retired, its IPv4 address is returned to the dynamic pool and becomes available for reuse by future mappings.

# Deployment Architectures

## CHARON alone

       IPv4 Island
   +----------------+
   |   IPv4 Hosts   |
   |                |
   |   GW: CHARON   |
   +--------+-------+
            |
            |
     +------+------+
     |   CHARON    |
     |  Translator |
     |   v4<->v6   |
     +------+------+
            |
            |
     +------+------+
     |  IPv6 Only  |
     |  Internet   |
     +-------------+

In this deployment, CHARON operates as the default gateway for the IPv4 island. The internal IPv4 prefix (e.g. 192.168.0.0/24) is mapped to an IPv6 prefix (e.g., 2001:db8:4600::/120) using stateless translation, while dynamic mappings are created as required for external IPv6 destinations.

All traffic from IPv4 hosts traverses the translator. As no native IPv4 upstream connectivity exists, communication is limited to destinations reachable over IPv6. IPv4-only external destinations are not reachable in this architecture.

## CHARON with native IPv4

       IPv4 Island
   +-----------------+
   |   IPv4 Hosts    |
   |                 |
   |   GW: Router    |
   +--------+--------+
            |
            |
     +------+------+
     | IPv4 Router |
     +--+-------+--+
        |       |
        |       | Dynamic Pool (routed)
        |       | e.g. 10.0.0.0/8
        |       |
        |   +---+--------+
        |   |   CHARON   |
        |   | Translator |
        |   +---+--------+
        |       |
        |       |
  +-----+--+  +-+------+
  |  IPv4  |  |  IPv6  |
  |Internet|  |Internet|
  +--------+  +--------+

In this deployment, the IPv4 island uses a conventional IPv4 router as its default gateway, providing native connectivity to the IPv4 Internet.

CHARON is deployed as a separate translator and is reachable via a dedicated IPv4 prefix (e.g., 10.0.0.0/8) that is routed from the IPv4 router to the translator. This prefix serves as the dynamic mapping pool.

Traffic destined for synthesized IPv4 addresses within the dynamic pool is routed to CHARON and translated to IPv6. All other traffic follows the native IPv4 path. This enables simultaneous access to both IPv4-only and IPv6-only destinations, using translation only when required.

The CHARON translator and IPv4 Router may be virtual functions in the same router. 

## CHARON with 464XLAT

       IPv4 Island
   +----------------+
   |   IPv4 Hosts   |
   |                |
   |   GW: CHARON   |
   +--------+-------+
            |
            |
     +------+------+
     |   CHARON    |
     |  Translator |
     |   v4<->v6   |
     +------+------+
        |       |
        |       | RFC6052 Prefix (routed)
        |       | e.g. 64:ff9b::/96
        |       |
        |   +---+--------+
        |   |    PLAT    |
        |   | Translator |
        |   +---+--------+
        |       |
        |       |
  +-----+--+  +-+------+
  |  IPv6  |  |  IPv4  |
  |Internet|  |Internet|
  +--------+  +--------+

In this deployment, CHARON operates as the default gateway for the IPv4 island. The internal IPv4 prefix (e.g. 192.168.0.0/24) is mapped to an IPv6 prefix (e.g., 2001:db8:4600::/120) using stateless translation, while dynamic mappings are created as required for external IPv6 destinations.

For packets which are not translated via a dynamic mapping, CHARON translates IPv4 packets using RFC6052-based encoded addresses (e.g. 64:ff9b::/96), acting as a CLAT in a 464XLAT architecture. 

All traffic from IPv4 hosts traverses the translator. This enables simultaneous access to both IPv4-only and IPv6-only destinations, using translation for all packets.

# Security Considerations

TODO Security

-Mapping range is NOT private address space


# IANA Considerations

This document has no IANA actions.


--- back

# Acknowledgments
{:numbered="false"}

TODO acknowledge.