<font color="red" size=5>This document is for informational purposes only</font>

```diff
-- This document is out of date and some content is for informational purposes only
```

1. ~~Modify the hosts file of the host~~
2. ~~Modify the container hostname value and image information~~
3. ~~It may cause watchtower to fail to update normally~~

## Nginx

[for reference only](https://github.com/XRSec/AUDCH/issues/1)

```bash
# Linux
systemctl disable --now systemd-resolved
systemctl stop systemd-resolved
mv /etc/resolv.conf /etc/resolv.conf.bak
```

### Run audch in docker

```bash
docker run -itd --name audch \
  --restart=always \
  -v /opt/hosts:/hosts \
  -v /var/run/docker.sock:/var/run/docker.sock \
  xrsec/audch
```

### Run dnsmasq in Docker

```bash
docker run -itd \
    --name dnsmasq \
    --restart always \
    -p 53:53 \
    -p 53:53/udp \
    -p 8080:80 \
    -v /opt/hosts:/hosts \
    -e "HTTP_USER=admin" \
    -e "HTTP_PASS=123456" \
    xrsec/dnsmasq
```

### Get dokcer0 IPAddress

```bash
docker network inspect --format='{{range .IPAM.Config}}{{.Gateway}}{{end}}' bridge
```

### Change Docker Config

> /etc/docker/daemon.json
>
> docker0 IPAddress: 172.17.0.1

```json
{
  "dns": [
    "172.17.0.1",
    "8.8.8.8",
    "8.8.4.4"
  ]
}
```

### Change DNS config

```yaml
# /etc/resolv.conf
nameserver 172.17.0.1 # docker0 IPAddress
nameserver 8.8.8.8
nameserver 8.8.4.4
nameserver 223.5.5.5
nameserver 223.6.6.6
nameserver 114.114.114.114

```

### Restart Docker Service

```bash
systemctl restart docker.service
```

### Check DNSServer

```bash
nslookup dnsmasq.docker.shared
```

```bash
docker logs -f audch
```
