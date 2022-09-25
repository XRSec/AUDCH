FROM golang:alpine AS builder
ARG buildTime
ARG versionData
ARG commitId
ARG author
COPY . /audch
WORKDIR /audch
RUN CGO_ENABLED=0 go build -a -ldflags "-extldflags -static -X main.buildTime=${buildTime} -X main.versionData=${versionData} -X main.commitId=${commitId} -X main.author=${XRSec}" -o audch .

FROM scratch
LABEL maintainer="xrsec"
LABEL mail="Jalapeno1868@outlook.com"
LABEL Github="https://github.com/XRSec/AUDCH"
LABEL org.opencontainers.image.source="https://github.com/XRSec/AUDCH"
LABEL org.opencontainers.image.title="AUDCH"
COPY --from=builder /audch/audch /audch
ENTRYPOINT ["/audch", "-f", "/hosts"]