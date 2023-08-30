# Simple SSL Certificate Monitoring Tools

This is working concept at this moment.

## Example of usage

Run manual check and display failed entries or expires in 7 days:

    /opt/bin/crt-check -file /opt/etc/crt-hosts.yml -days 7

Run manual check with enabled IPv6 support and display all entries:

    /opt/bin/crt-check -file /opt/etc/crt-hosts.yml -6

Launch Prometheus exporter on port 2112

    /opt/bin/crt-mon -file /opt/etc/crt-hosts.yml -port 2112

## TODO
- Install script
- Better print unresolvable domains error by crt-check
