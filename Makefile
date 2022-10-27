all:
	@docker rm -f nginx audch
	@docker run -itd --name nginx --rm nginx
	@docker build -f Dockerfile.test -t xrsec/audch .
	@docker run -it --name audch --rm -p 53:53 -p 53:53/udp -v /opt/xrsec/docker/dnsmasq/hosts:/hosts -v /var/run/docker.sock:/var/run/docker.sock xrsec/audch