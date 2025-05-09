# About
Minimal Docker secrets plugin for [Passwordstate](https://www.clickstudios.com.au/passwordstate.aspx)

Inspired by [Docker/Vault secrets plugin](https://gitlab.com/sirlatrom/docker-secretprovider-plugin-vault) and [documentation](https://blog.sunekeller.dk/2019/04/vault-swarm-plugin-poc/) related to it.

TODO:
* Password rotation without restarting containers

# Usage
## Plugin installation
```bash
docker plugin install \
    --alias pwdstate \
    --grant-all-permissions \
    ollijanatuinen/docker-secretprovider-plugin-passwordstate:v0.2 \
    PASSWORDSTATE_BASE_URL="https://passwordstate/api" \
    PASSWORDSTATE_API_KEY="<api key>" \
    PASSWORDSTATE_LIST_ID="123"
```

## Deploy test container and verify that secret is visible
**NOTE!!!** Swarm does not check if secret exist in backend shen calling `docker secret create`.
It is only requested when task using secret is allocated.
```bash
$ docker secret create --driver pwdstate test1
f6tudm7i13hz47bg8dvjl7tam

$ docker service create --name test --secret test1 nginx
snk5zxolssjum9kecdpldtdu0
overall progress: 1 out of 1 tasks
1/1: running   [==================================================>]
verify: Service snk5zxolssjum9kecdpldtdu0 converged

$ docker ps
CONTAINER ID   IMAGE          COMMAND                  CREATED              STATUS              PORTS     NAMES
2ffaaa947950   nginx:latest   "/docker-entrypoint.…"   About a minute ago   Up About a minute   80/tcp    test.1.5m1mge0rf2pwnmdirejr3v81z

$ docker exec -it test.1.5m1mge0rf2pwnmdirejr3v81z cat /run/secrets/test1
nNDuUPKfr8LQQgHY2Faz#-MP8b+t
```
