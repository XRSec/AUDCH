## README

- Added automatic update hostname function, which may cause watchtower to fail to update normally
- Modify the image and add the original image information to Labels.AUDCH_IMAGE
- Fixed NGINX not parsing HOSTS

## USE

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
