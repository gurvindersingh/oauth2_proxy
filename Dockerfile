FROM alpine:3.4
MAINTAINER Gurvinder Singh <gurvinder.singh@uninett.no>
RUN apk update && apk add ca-certificates
ADD config /config
ADD oauth2_proxy /oauth2_proxy
ENTRYPOINT ["/oauth2_proxy"]
CMD ["-config", "config/config.toml"]