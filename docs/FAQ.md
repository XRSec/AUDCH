1. ~~Modify the hosts file of the host~~
2. ~~Modify the container hostname value and image information~~
3. ~~It may cause watchtower to fail to update normally~~

## Nginx

```bash
docker run -itd \
    --name dnsmasq \
    --restart always \
    -p 8082:8080 \
    -v /docker/dnsmasq/dnsmasq.conf:/etc/dnsmasq.conf \
    -v /etc/hosts:/hosts \
    -e "HTTP_USER=admin" \
    -e "HTTP_PASS=123456" \
    jpillora/dnsmasq
    
docker inspect --format='{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' dnsmasq
```
```json
# /etc/docker/daemon.json
        
{
  "dns": ["172.17.0.3"]
}
```

```bash
systemctl restart docker.service
```