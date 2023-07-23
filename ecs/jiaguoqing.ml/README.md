# Setup
## Create user and group
```shell
useradd
```

## Config rsyslog
Create directory for log file
```shell
mkdir /var/log/brother
chown syslog:syslog /var/log/brother
```
Then add config file /etc/rsyslog.d/brother.conf
```shell
if $programname == 'brother' then /var/log/brother/output.log
& stop
```
Restart rsyslog
```shell
systemctl restart rsyslog.service
```

## Config caddy
## Config systemd

## Reference
- https://luci7.medium.com/golang-running-a-go-binary-as-a-systemd-service-on-ubuntu-18-04-in-10-minutes-without-docker-e5a1e933bb7e
