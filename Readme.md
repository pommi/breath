# breath

Route specific domain names to be selectively redirected to openvpn gateway.

**breath** takes list of domain names as config, resolves them
and pushes routes for these IP addresses to be redirected via OpenVPN gateway (`tun0`).

It was developed to keep routes up to date with changes in DNS and is tested to
work flawlessly with **openvpn** client mode.

### Development status

Ready to use

### OS Support

Depends on **netlink** support in the system.

- Linux
- FreeBSD

## Config file

YAML syntax is used for config file.

File name is hard-coded, `breath.yml`. **It is required to create config file manually.**

Here is an example.

```yml
version: "1"
target:
  gateway: 10.8.0.1
  name: tun0
sources:
  - interval: 5m
    domains:
      - google.com
      - google.co.uk
  - interval: 48h
    domains:
      - google.fr
default_resolver:
  nameservers: [ 8.8.8.8, 8.8.4.4 ]
  on_failure: hold
```

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

`sudo /home/ubuntu/bin/breath`

# How it works

Route management and interface state tracking is implemented using
[vishvananda/netlink](https://github.com/vishvananda/netlink) package.

## TODO List
- [ ] cache initial resolution to bootstrap restarts
- [ ] track vpn/link status to remove routes when VPN is not ready
- [ ] systemd daemon mode support for without-docker (tweak for logging and add sample unit file)
- [ ] support for `auto` interval
