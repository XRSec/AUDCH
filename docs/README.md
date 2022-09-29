## AUDCH

Auto Update Docker Container TO User Hosts

> Get IP from docker list every 5 minutes and add hosts alias
>
> ~~<font color="red"> We will recreate the container, you need to know and be clear about this </font>~~

## Notice

> Requires read and write permissions to `/etc/hosts`

- ? Docker `--cap-add`
- ? Docker `--privileged`
- ? 5 minutes

[![nginx](/docs/nginx.png)](/docs/EXAMPLE.md)

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
```

#### <font color="green">Only update the hosts file</font>

```bash
docker run -itd --name audch \
  --restart=always \
  -v /etc/hosts:/hosts \
  -v /var/run/docker.sock:/var/run/docker.sock \
  xrsec/audch
```

#### ~~<font color="red">Recreate container [ in development ]</font>~~

```bash
docker run -itd --name audch \
  --restart=always \
  -e recreate=true \
  -v /etc/hosts:/hosts \
  -v /var/run/docker.sock:/var/run/docker.sock \
  xrsec/audch
```

## FAQ

- ~~`docker commit -m "AUDCH Update hostname" CONTAINER CONTAINER/Image`~~
- ~~edit CONTAINER json file~~
- dnsmasq
