# About
Minimal Docker secrets plugin for [Passwordstate](https://www.clickstudios.com.au/passwordstate.aspx)

Inspired by [Docker/Vault secrets plugin](https://gitlab.com/sirlatrom/docker-secretprovider-plugin-vault) and [documentation](https://blog.sunekeller.dk/2019/04/vault-swarm-plugin-poc/) related to it.

# Usage
## Plugin installation
```bash
docker plugin install \
    --alias pwdstate \
    --grant-all-permissions \
    ollijanatuinen/docker-secretprovider-plugin-passwordstate:v1.0 \
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

## Test secret rotation
1. Update password in Passwordstate
2. Scale service to two replicas and see that they are using different passwords
```bash
$ docker service scale test=2
test scaled to 2
overall progress: 2 out of 2 tasks
1/2: running   [==================================================>]
2/2: running   [==================================================>]
verify: Service test converged

$ docker ps
CONTAINER ID   IMAGE          COMMAND                  CREATED              STATUS              PORTS     NAMES
2cbc77de9cd6   nginx:latest   "/docker-entrypoint.…"   6 minutes ago        Up 6 minutes        80/tcp    test.1.2m9u2x6ymx3jky4jxn849mvar
6379778ff1fa   nginx:latest   "/docker-entrypoint.…"   About a minute ago   Up About a minute   80/tcp    test.2.46hjx3r73eyeiqtjjy6ca0nr6

$ docker exec -it test.1.2m9u2x6ymx3jky4jxn849mvar cat /run/secrets/test1
nNDuUPKfr8LQQgHY2Faz#-MP8b+t

$ docker exec -it test.2.46hjx3r73eyeiqtjjy6ca0nr6 cat /run/secrets/test1
6wBKTyhvp3k5k_5yU2ituLmezgfA3
```
**NOTE!!!** This means that to finalize secret rotation, you must restart all containers.
Also if you are using secret which requires username+password pair instead of API key, the you should store them both to Password field.

# Troubleshooting
```bash
cat /var/lib/docker/plugins/<plugin id>/rootfs/pwdstate.log
time="2025-05-09T05:16:45Z" level=info msg="Starting Docker secrets plugin"
time="2025-05-09T05:17:03Z" level=info msg="Secret 'test1' requested by test (test.1.2m9u2x6ymx3jky4jxn849mvar)"
time="2025-05-09T05:17:31Z" level=info msg="Secret 'test1' requested by test (test.2.46hjx3r73eyeiqtjjy6ca0nr6)"
```
