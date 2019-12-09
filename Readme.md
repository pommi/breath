# breath

Route specific domain names to be selectively redirected to openvpn gateway.

## How it works

**breath** uses list of configured domain names, resolves them
and pushes routes for these IP addresses to be redirected via OpenVPN gateway (`tun0`).

It was developed to keep routes up to date with changes in DNS and tested to
work flawlessly with **openvpn** client mode.

### OS Support

Depends on **netlink** support in the system.

- Linux
- FreeBSD

## Config file

YAML syntax is used, please ensure that your pad your sections with **2 or 4 spaces, not tabs**.
File name is hard-coded, `breath.yml`. **It is required to create config file manually.**

Here is an example.

```yml
version: "1"
target:
  name: tun0            # vpn interface
  gateway: 10.8.0.1/2   # vpn server
sources:
  - interval: 5m
    domains:
      - google.com
      - google.co.uk
  - interval: 48h  #NOTE: "days" not supported, specify hours (72h instead of 3d)
    domains:
      - google.fr
default_resolver:
  nameservers: [ 8.8.8.8, 8.8.4.4 ]
  on_failure: hold # hold / drop
```

## Running

Run this on VPN client host. Added routes are automatically removed if VPN client is disconnected.

### With Docker

1. Clone the repo
2. Inside cloned directory, do `docker-compose build`
3. Write a config file (see above) and save it as **`breath.yml`**

Create persistent container that will do the job:
```sh
docker run -d \
  --name breath \
  -v $(pwd)/breath.yml:/home/breath/app/breath.yml \
  --restart unless-stopped \
  --cap_add NET_ADMIN
  breath
```


### Without Docker

1. Clone the repo
2. Write config, save as **`breath.yml`** inside cloned directory
3. Install `make` and `go` 1.9+ for your distribution (ensure `go version` prints the version after install)
4. Build utility with command: `make build` (binary is saved to `../bin`)
4. To start breath worker use command:

`/home/ubuntu/bin/breath`

# How it works

Route management and interface state tracking is implemented using
[vishvananda/netlink](https://github.com/vishvananda/netlink) package.

When the breath starts, it waits for interface activation (say, `tun0`),
and then resolves all domain names. For every resolved IP address, it adds a route
so that system will use redirect traffic to that IP address.

## TODO

- systemd daemon mode support for without-docker rolling (+logging)
- support for `auto` interval to adapt to the NS SOA record expiry from DNS responses in a group
- cache initial resolution to bootstrap restarts
