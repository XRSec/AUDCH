## AUDCH

Auto Update Docker Container TO User Hosts

<font color="green">From Docker List ( cron: 5 min ) </font>

## Notice

- Please add DNS address manually
- For non-docker containers, please install dnsmasq

[![nginx](/docs/images/nginx.png)](/docs/EXAMPLE.md)

## Usage

```bash
Audch -h # help
Audch -c /etc/hosts # don't use dnsmasq

Audch --install # shell auto-completion install
Audch --uninstall # shell auto-completion uninstall

Audch -c /opt/audch/hosts -e dnsmasq -c /opt/audch/hosts -p 80 -- dnsmasq --no-daemon # use dnsmasq
Audch -c /opt/audch/hosts -e dnsmasq -c /opt/audch/hosts -c /opt/dnsmasq/dnsmasq.conf -p 80 -- dnsmasq --no-daemon # use dnsmasq

... # more
```

![dns](/docs/images/dns2.png)

### Docker

```diff
++ Use dnsmasq by default
```

```bash
mkdir audch
touch audch/hosts
docker run -itd \
  --name audch \
  --restart=always \
  -p 8080:80 \
  -p 53:53 \
  -p 53:53/udp \
  -e "HTTP_USER=admin" \
  -e "HTTP_PASS=123456" \
  -v ./audch/hosts:/hosts \
  -v /var/run/docker.sock:/var/run/docker.sock \
  xrsec/audch
```

## FAQ

- [ ] ~~regester service~~
- [x] [fix nginx resolv](https://github.com/XRSec/AUDCH/issues/1)
