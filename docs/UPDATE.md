## README

- Added automatic update hostname function, which may cause watchtower to fail to update normally
- Modify the image and add the original image information to Labels.AUDCH_IMAGE
- Fixed NGINX not parsing HOSTS

## USE

```bash
sudo chmod 666 /etc/hosts

docker run -itd --name audch \
  --restart=always \
  -v /etc/hosts:/hosts \
  -v /var/run/docker.sock:/var/run/docker.sock \
  xrsec/audch
  ```
