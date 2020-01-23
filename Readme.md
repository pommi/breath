# breath

Route specific domain names to be selectively redirected to openvpn gateway.

**breath** takes list of domain names as config, resolves them
and pushes routes for these IP addresses to be redirected via OpenVPN gateway (`tun0`).

It was developed to keep routes up to date with changes in DNS and is tested to
work flawlessly with **openvpn** client mode.

### Development status

It works as expected (see TODO List below)

### OS Support

Depends on **netlink** socket API support in the system.

- Linux
- BSD

## Config file

YAML syntax is used for config file.

File name is hard-coded, `breath.yml`. **It is required to create config file manually.**

Here is an example. It will ensure specific domain names are redirected.

```yml
version: "1"

target:
  name: tun0
  gateway: 10.8.0.1
  metric: 10

default_resolver:
  nameservers: [ 8.8.8.8, 8.8.4.4 ]
  on_failure: hold

sources:
  - interval: 5m
    domains:
      - google.com
      - google.co.uk
      - google.fr
```

## Run

### With Docker

1. Clone the repo
2. Inside cloned directory, do `docker-compose build`
3. Write a config file (see above) and save it as **`breath.yml`**

Create and start persistent container that will do the job:
```sh
docker run -d \
  --name breath \
  --net host \
  --cap-add NET_ADMIN \
  -v $(pwd)/breath.yml:/root/app/breath.yml \
  --restart unless-stopped \
  breath
```

### Without Docker

1. Clone the repo as `/$HOME/breath`
2. Write config, save as **`breath.yml`** inside cloned directory
3. Install `make` and `go` 1.9+ for your distribution (ensure `go version` prints the version after install)
4. Build utility with command: `make build`. Binary is in `../bin`
4. To start breath worker use command:

```sh
sudo /$HOME/bin/breath
```

# License

BSD 3-Clause License

# How it works

Route management is done through netlink kernel API implemented by
[vishvananda/netlink](https://github.com/vishvananda/netlink) Go package.

## TODO List
- [x] add and remove routes, auto-update routes with interval
- [ ] track link status. If link is down, sleep. If link goes up, re-add routes
- [ ] cache initial resolution to bootstrap restarts
- [ ] systemd daemon mode support for without-docker (tweak for logging and add sample unit file)
- [ ] support for `auto` interval
- [ ] add DNS-over-HTTPs support with force/try mode for resolvers
