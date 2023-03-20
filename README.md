# CMS audit

Simple audit

## Install

```yaml
version: '3.7'
services:
    audit:
        image: janatjak/cmsaudit
        environment:
            TOKEN: xxx
            DATABASE_DSN: 'host=127.0.0.1 user=root password=pass dbname=audit sslmode=disable'
```

Create docker stack
```shell
docker stack deploy -c docker-stack.yml audit
```