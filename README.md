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
    ollijanatuinen/docker-secretprovider-plugin-passwordstate:v0.1 \
    PASSWORDSTATE_BASE_URL="https://passwordstate/api" \
    PASSWORDSTATE_API_KEY="<api key>"
```

## Deploy test container and verify that secret is visible
```bash
$ docker stack deploy -c docker-compose.yml --detach=false test
Creating network test_default
Creating secret test1
Creating service test_app
overall progress: 1 out of 1 tasks
1/1: running   [==================================================>]
verify: Service 6t2sqgnkf74lujpeoeq6wpxjo converged

$ docker secret ls
ID                          NAME      DRIVER     CREATED          UPDATED
zmn2431u8a8z1s9xr25wjtv47   test1     pwdstate   26 seconds ago   26 seconds ago

$ docker ps
CONTAINER ID   IMAGE          COMMAND                  CREATED          STATUS          PORTS     NAMES
cbc4828ddd57   nginx:latest   "/docker-entrypoint.…"   32 seconds ago   Up 32 seconds   80/tcp    test_app.1.lrt2pb5i5i06oif90kncj9n23

$ docker exec -it test_app.1.lrt2pb5i5i06oif90kncj9n23 cat /run/secrets/test1
7wf98+N3GJgQBnrqS63V6EqdXYu
```
