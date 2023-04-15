## README

- Remove dnsmasq and use a custom dns server, which is more portable

## USE

```bash
mkdir audch
touch audch/hosts
docker run -itd \
  --name audch \
  --restart=always \
  -p 53:53/udp \
  -v ./audch/hosts:/hosts \
  -v /var/run/docker.sock:/var/run/docker.sock \
  xrsec/audch
  ```

More details can be found in the [README](/docs/README.md) AND [FAQ](/docs/FAQ.md)
