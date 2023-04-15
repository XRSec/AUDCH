FROM golang:1.20-alpine AS builder
ARG versionData
COPY . /audch
WORKDIR /audch
RUN touch /hosts \
    && CGO_ENABLED=0 go build -a -ldflags "-extldflags -static -X main.versionData=${versionData}" -o audch .

FROM scratch
LABEL maintainer="xrsec"
LABEL mail="Jalapeno1868@outlook.com"
LABEL Github="https://github.com/XRSec/AUDCH"
LABEL org.opencontainers.image.source="https://github.com/XRSec/AUDCH"
LABEL org.opencontainers.image.title="AUDCH"
ENV TZ Asia/Shanghai

COPY --from=builder /audch/audch /Audch
COPY --from=builder /hosts /hosts
EXPOSE 53/udp

ENTRYPOINT ["/Audch", "-c", "/hosts"]
