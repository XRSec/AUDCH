# AUDCH

Auto Update Docker Container TO User Hosts

> Get IP from docker list every 5 minutes and add hosts alias

## Notice

> Requires read and write permissions to `/etc/hosts`

- ? Docker `--cap-add`
- ? Docker `--privileged`
- ? 5 minutes

## Usage

```bash
Usage of audch
  -f string
        hosts filepath (default "/etc/hosts")
  -v    show version
```

### Screen

```bash
sudo chmod 666 /etc/hosts

screen -UR audch
./audch

# ctrl + a + d
# screen -Ur audch
```

### Docker

```bash
sudo chmod 666 /etc/hosts

docker run -itd --name audch \
  --restart=always \
  -v /etc/hosts:/hosts \
  -v /var/run/docker.sock:/var/run/docker.sock \
  xrsec/audch
```

## FAQ

- ~~`docker commit -m "AUDCH Update hostname" CONTAINER CONTAINER/Image`~~
- ~~edit CONTAINER json file~~
- dnsmasq