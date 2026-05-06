# Styx
The cursed river that helps your IPv4-only clients exit the underworld to the IPv6-only future


## Basically how it works:

$\color{#789922}\text{>have old ass computer which can only do IPv4}$\
$\color{#789922}\text{>maybe am vintage collector and like SGI}$ $\color{#789922}\text{workstations, they are beautiful}$\
$\color{#789922}\text{>the world has moved to IPv6, can't browse}$ $\color{#789922}\text{4chan on your old machine}$\
$\color{#789922}\text{>old ass computer sends A-record DNS query}$ $\color{#789922}\text{to Styx}$\
$\color{#789922}\text{>styx finds glorious AAAA-record, and uses it}$ $\color{#789922}\text{to plot a course out of the underworld}$\
$\color{#789922}\text{>styx returns the entry point of this course as a}$ $\color{#789922}\text{valid IPv4 address in your chosen translation subnet}$\
$\color{#789922}\text{>tayga ferries packets from the underworld}$ $\color{#789922}\text{to the modern world and back along this course}$\
$\color{#789922}\text{>client browses 4chan from IRIX happily}$


## How it actually works:
- You setup `styx` as a CLAT (client-side translator), to provide native IPV4 support for your legacy clients in an IPv6 world
- `styx` provides DNS services to clients over IPv4, and also launches `tayga` to provide v4->v6 packet translation
- When the client requests an A record, and an AAAA record is available, `styx` pulls an unused IPv4 out of the translation pool (Default pool is `10.0.0.0/8`, but you can use any IPv4 CIDR range) and returns this address to the client
- At the same time, a new mapping is added in `tayga` which maps this new IPv4 to the result IPv6 address
- Example: you browse to `google.com` (returns AAAA `2a00:1450:4026:801::200e`). `styx` returns `10.0.0.1` to the client, then adds `map 10.0.0.1 2a200:1450:4026:801::200e` to `tayga`. 
- This mapping session persists indefinitely, and will be re-used if any subsequent DNS queries result in IPv6 address `2a00:1450:4026:801::200e`
- Currently these entries never time out, maybe I should fix that

## Installation
You need the latest `main` of `tayga` as well as this repository.

Instructions on Debian (you may need sudo, idk):
```sh
#install dependencies
apt install git golang-go build-essential
#make + install tayga
git clone https://github.com/apalrd/tayga
cd tayga
make
make install
#make + install styx
cd ..
git clone https://github.com/apalrd/styx46
cd styx46
make
#copy config example to running config
cp styx.yaml.example styx.yaml
#edit the config for your setup

#run it
./styx
```

`styx` assumes that IPv6 connectivity is already working, possibly including nat64 (default prefix is `64:ff9b::/96`). nat64 is not strictly needed if all of your desired services support IPv6 already. You must have an IPv6 default gateway and must *not* have an IPv4 default gateway (`styx` will become the local gateway). `styx` will enable IP forwarding, and proxy ND for each of the addresses behind the CLAT. This may disable `accept_ra` depending on your distribution, in which case you may need to re-enable it: `echo "2" > /proc/sys/net/ipv6/conf/eth0/accept_ra` (if your upstream connectivity is via `eth0`).


### disclaimer: this was made in two days and only tested by doing apt update + apt upgrade, and pinging google. I have never tested it with more than ~5 mapping entries at once. You should not trust styx in a production network
