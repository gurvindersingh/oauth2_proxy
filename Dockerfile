FROM scratch
MAINTAINER Gurvinder Singh <gurvinder.singh@uninett.no>
ADD config /config
ADD oauth2_proxy /oauth2_proxy
ENTRYPOINT ["/oauth2_proxy"]
CMD ["-config", "config/config.toml"]