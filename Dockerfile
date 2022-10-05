# FROM golang:alpine AS goBuilder
# ARG versionData
# COPY . /audch
# WORKDIR /audch
# RUN CGO_ENABLED=0 go build -a -ldflags "-extldflags -static -X main.versionData=${versionData}" -o audch .

# FROM gcc AS gccBuilder
# RUN curl https://thekelleys.org.uk/dnsmasq/dnsmasq-2.87.tar.gz -o - | tar -xz \
#     && cd dnsmasq-* \
#     && make \
#     && make install DESTDIR=/dnsmasq \
#     && set -ex \
#     && tar -cphf so.tar $(ldd -v /dnsmasq/usr/local/sbin/dnsmasq | awk '{print $4}' | grep so | sort | uniq | tr '\n' ' ') \
#     && tar -xf so.tar -C /dnsmasq

# FROM scratch
# LABEL maintainer="xrsec"
# LABEL mail="Jalapeno1868@outlook.com"
# LABEL Github="https://github.com/XRSec/AUDCH"
# LABEL org.opencontainers.image.source="https://github.com/XRSec/AUDCH"
# LABEL org.opencontainers.image.title="AUDCH"
# COPY --from=goBuilder /audch/audch /Audch
# COPY --from=gccBuilder /dnsmasq /
# USER root
# ENTRYPOINT ["/Audch", "-c", "/hosts", "-e", "dnsmasq", "-c", "/host", "-c", "/etc/dnsmasq.conf", "-p", "80", "--", "dnsmasq", "-u rot", "--no-daemon"]

FROM alpine
LABEL maintainer="xrsec"
LABEL mail="Jalapeno1868@outlook.com"
LABEL Github="https://github.com/XRSec/AUDCH"
LABEL org.opencontainers.image.source="https://github.com/XRSec/AUDCH"
LABEL org.opencontainers.image.title="AUDCH"
ARG TARGETARCH
COPY bin/Audch-linux-${TARGETARCH} /Audch

RUN apk update \
	&& apk --no-cache add dnsmasq \
    && apk add --no-cache --virtual .build-deps curl \
    && curl -sL https://raw.githubusercontent.com/imp/dnsmasq/master/dnsmasq.conf.example -o /etc/dnsmasq.conf \
    && apk del .build-deps \
	&& echo -e "addn-hosts=/hosts\nlog-queries\nno-hosts\nall-servers" >> /etc/dnsmasq.conf


EXPOSE 53/udp 53/tcp 80

ENV PORT=80 HTTP_USER="" HTTP_PASS=""

ENTRYPOINT ["/Audch", "-c", "/hosts", "-e", "dnsmasq", "-c", "/host", "-c", "/etc/dnsmasq.conf", "-p", "80", "--", "dnsmasq", "--no-daemon"]