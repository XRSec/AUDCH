1. ~~Modify the hosts file of the host~~
2. ~~Modify the container hostname value and image information~~
3. ~~It may cause watchtower to fail to update normally~~

## Nginx

[for reference only](https://github.com/XRSec/AUDCH/issues/1)

```bash
systemctl disable --now systemd-resolved
systemctl stop systemd-resolved
mv /etc/resolv.conf /etc/resolv.conf.bak

docker run -itd \
    --name dnsmasq \
    --restart always \
    -p 53:53 \
    -p 53:53/udp \
    -p 35012:35012 \
    -v /etc/hosts:/hosts \
    -e "HTTP_USER=admin" \
    -e "HTTP_PASS=123456" \
    xrsec/dnsmasq
```

```json
// /etc/docker/daemon.json

{
  "dns": [
    "172.17.0.1",
    "8.8.8.8",
    "8.8.4.4"
  ],
}
```

```yaml
# /etc/resolv.conf
nameserver 172.17.0.1
nameserver 8.8.8.8
nameserver 8.8.4.4
nameserver 223.5.5.5
nameserver 223.6.6.6
nameserver 114.114.114.114

```

```bash
systemctl restart docker.service
```
