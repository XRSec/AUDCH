## AUDCH

<details>
  <summary>Auto Update Docker Container TO User Hosts.</summary>
    <p>Suppose your company uses docker as the development environment, and suddenly one day the server breaks down, the server needs to be migrated, and the IP address needs to be reconfigured again.</p>
    <p></p>
    <p>But I believe that if you are smart, you will definitely use docker-compose to configure an IP address for each container, but can you guarantee that you can remember this IP address every time?</p>
    <p></p>
    <p>If you plan to use adminer to query mysql data, are you going to create a new container in the existing docker-compose to achieve the purpose of query? Or use docker run --link mysql=mysql ?</p>
    <p></p>
    <p>No no no, it's too cumbersome, you only need 30S to do it once and for all.</p>
    <p></p>
    <p>Audch is a docker tool, smaller and simpler than coredns. It will be automatically added to the dns system when creating a container, and will be automatically deleted when deleting a container. The alias configuration method is simpler, that is, container name + .docker.shared.audch , and will automatically connect containers on different network segments, such as containers created by docker-compose, to the bridge network, so that all containers are on the same network segment and can access each other.</p>
    <p></p>
    <p>A series of operations avoid port forwarding problems, which are safer and less troublesome.</p>
    <p></p>
    <p>tips: Since macOS has some difficult-to-operate problems, macOS users are not recommended to modify the dns according to the operation, because even if the dns is modified, it cannot be called, because macOS involves the mdns problem, and even if the mdns is solved, the docker container does not allow the host to directly Access can only be connected by port forwarding. It can be solved by other methods, but it is not necessary</p>
</details>

<details>
  <summary>为每个Dockerr容器生成一个可互相访问的域名</summary>
    <p>假设你们公司使用 docker 作为开发环境，突然有一天服务器坏了，需要迁移服务器，又需要重新配置一遍IP地址。</p>
    <p></p>
    <p>但是相信聪明的你肯定会使用 docker-compose 为每个容器配置一个IP地址，但是你能保证你每次都能记住这个IP地址吗？</p>
    <p></p>
    <p>如果你准备使用adminer 查询mysql数据，难道你准备在已有的docker-compose 中新建一个容器来达到查询的目的吗？或者是使用 docker run --link mysql=mysql ? </p>
    <p></p>
    <p>不不不，太繁琐了，你只需30S就能一劳永逸。</p>
    <p></p>
    <p>audch 属于docker 小工具，比 coredns 更小，更简单，创建容器的时候会自动添加到 dns 系统里面，删除容器的时候会自动删除，别名配置方法更加简单，就是 容器名+ .docker.shared. audch，并且还会自动把不同网段的容器，比如 docker-compose 创建的容器 连接到 bridge 网络，这样所有的容器都在同一网段，能够互相访问。</p>
    <p></p>
    <p>一系列操作，避免了端口转发问题，更安全，更省事，</p>
    <p></p>
    <p>tips： 由于macOS 具有一些难以操作的问题，所以不建议macOS用户根据操作修改dns，因为即使修改了dns也无法调用，因为macOS涉及到了mdns 问题，且就算mdns 解决了，docker 容器不允许宿主机直接访问，只能使用端口转发的方式连接，可以用其他方法解决，但是没必要</p>
</details>

## Notice

- Please add DNS address manually

[![nginx](/docs/images/nginx.png)](/docs/EXAMPLE.md)

## Usage

```bash
Audch -h # help
Audch -c /etc/hosts # don't use dnsmasq

Audch --install # shell auto-completion install
Audch --uninstall # shell auto-completion uninstall

Audch -c /opt/audch/hosts

... # more
```

![dns](/docs/images/dns2.png)

### Docker

```bash
mkdir audch

touch audch/hosts

docker run -itd \
  --name audch \
  --restart=always \
  -p 8080:80 \
  -p 53:53 \
  -p 53:53/udp \
  -v ./audch/hosts:/hosts \
  -v /var/run/docker.sock:/var/run/docker.sock \
  xrsec/audch
```

## FAQ

- [x] ~~regester service~~
- [x] [fix nginx resolv](https://github.com/XRSec/AUDCH/issues/1) ( but ↓
- [ ] [macOS can't resolv](https://stackoverflow.com/questions/76022499/the-dns-server-developed-by-golang-cannot-be-resolved) ( but ↓
- [ ] [macOS can't vist](https://github.com/docker/for-mac/issues/2670) ( so, The main container can be accessed normally）
