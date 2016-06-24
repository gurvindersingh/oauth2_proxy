oauth2_proxy
=================

### This project is fork of [Bitly Oauth2_proxy](https://github.com/bitly/oauth2_proxy)

A reverse proxy and static file server that provides authentication using Providers (Dataporten, Google, Github, and others)
to validate accounts and authorize based on the group memberships. More detailed information about options can be found on upstream project. Remember your application registered in auth provider **must** have **email** and **groups** scope.

## Architecture

![OAuth2 Proxy Architecture](https://cloud.githubusercontent.com/assets/45028/8027702/bd040b7a-0d6a-11e5-85b9-f8d953d04f39.png)

## Configuraiton

There is an example config file in config folder which explains few importants parameter to configure for dataporten. For all others, please look upstream documentation or as usual code :)

## Run

There is a _Dockerfile_ in the repo and a published docker image. To run using docker

```
docker run -it gurvin/oauth2_proxy:latest
```

To see the command line parameters which can be used in addition to config file

```
docker run -it gurvin/oauth2_proxy:latest -h
```