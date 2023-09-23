# Simple SSL Certificate Monitoring Tools

Simple SSL certificate expiration check tools:
 - **crt-check**: CLI check tool
 - **crt-mon**: HTTP server that's check expiration and expose as Prometheus metrics

For **crt-mon** is also available [ZABBIX template](https://github.com/xcdr/crt-mon/tree/main/install/zabbix).

## Installation

Download and extract release package from: https://github.com/xcdr/crt-mon/releases.

## Examples of usage

Run manual check and display failed entries or expires in 7 days:

    /opt/bin/crt-check -file /opt/etc/crt-hosts.yml -days 7

Run manual check with enabled IPv6 support and display all entries:

    /opt/bin/crt-check -file /opt/etc/crt-hosts.yml -6

Launch Prometheus exporter on port 2112:

    /opt/bin/crt-mon -file /opt/etc/crt-hosts.yml -port 2112

## TODO
- Install script
