# API Protector CLI


## Buidling CLI:
`make build`

## Prerequisites 

* Docker
* Docker Compose
* logrotate

setup docker to be executable by the user (Add to docker group) that is going to run this program.

## Config file to add to /etc/logrotate.d/:


docker_logs.config

```
/app/data/docker/*/logs/*.log {
    su [REPLACEUSERNAME] [REPLACEUSERGROUP]
    daily
	missingok
	rotate 10
	notifempty
    copytruncate
	create 0777
}
```

Then add to cron as required (Hourly or  Half Hourly etc)

`sudo logrotate docker_logs.conf`

The username and usergroup above should be the user that is going to write and read the log files. File permission can also be changed if required. 

The file `docker_logs.config` needs to be set to root user editable only. Else logrootate wont start.


## Example commands

### Validate

```bash
go run ./cmd --module validate --api-file /home/somelocation/openapi3.yml
``` 

### Deploy:

```bash
go run ./cmd --module deploy --api-file /home/somelocation/openapi3.yml --label examplelablename --listen-port 1321 --api-url http://172.18.0.1:8080 --request-mode monitor --health-port 1242`

```
### Undeploy:

```bash
go run ./cmd --module undeploy --label examplelablename
```

