# vpn-breath

Resolves domain names and pushes routes to redirect-gateway for resolved IPs only.


## How it works

**vpn-breath** continuously resolves configured list of desired domain names
and pushes routes to client's VPN interface (`tun0`) needed to
selectively redirect traffic to resolved IP addresses via the VPN gateway
(server).

It is written in Go.

## Config file

YAML syntax is used, please ensure that your pad your sections with **2 or 4 spaces, not tabs**.
File name is hard-coded, `breath.yml`. **It is required to create config file manually.**

Here is an example.

```yml
version: "1"
target:
  name: tun0
  gateway: 10.8.0.1
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
  on_fail: hold # hold / drop
```

## How to run

Run this on VPN client host. Added routes are atuomatically removed if VPN client is disconnected.

### With Docker

1. Clone the repo
2. Inside cloned directory, do `docker-compose build`
3. Write a config file (see above) and save it as **`breath.yml`**

Create persistent container that will do the job:
```sh
docker run -d \
  --name breath \
  -v $(pwd)/breath.yml:/home/vpn-breath/app/breath.yml \
  --restart unless-stopped \
  --cap_add NET_ADMIN
  vpn-breath
```


### Without Docker

1. Clone the repo
2. Write config, save as **`breath.yml`** inside cloned directory
3. Install `make` and `go` 1.9+ for your distribution (ensure `go version` prints the version after install)
4. Build utility with command: `make build`
4. To start vpn-breath worker use command:

`/home/ubuntu/bin/breath`

TODO:

- systemd daemon mode support for without-docker rolling (+logging)
- support for `auto` interval to adapt to the NS SOA record expiry from DNS responses in a group
