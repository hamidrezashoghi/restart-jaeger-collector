# restart-jaeger-collector

##### First create D-Bus connection

The `dbus` package connects to the [systemd D-Bus API](http://www.freedesktop.org/wiki/Software/systemd/dbus/) and lets you start, stop and introspect systemd units.
[API documentation][dbus-doc] is available online.

[dbus-doc]: https://pkg.go.dev/github.com/coreos/go-systemd/v22/dbus?tab=doc

Create `/etc/dbus-1/system-local.conf` that looks like this:

```
<!DOCTYPE busconfig PUBLIC
"-//freedesktop//DTD D-Bus Bus Configuration 1.0//EN"
"http://www.freedesktop.org/standards/dbus/1.0/busconfig.dtd">
<busconfig>
    <policy user="root">
        <allow eavesdrop="true"/>
        <allow eavesdrop="true" send_destination="*"/>
    </policy>
</busconfig>
```

##### Create jaeger-collector.service unitd

Create `/etc/systemd/system/jaeger-collector.service` that looks like this:
```
[Unit]
Description=Jaeger Collector
Requires=docker.service
Before=multi-user.target
After=docker.service
Wants=network-online.target

[Service]
Type=oneshot
WorkingDirectory=/opt/collector/
ExecStart=/usr/local/bin/docker-compose -f /opt/collector/collector.yaml up -d --remove-orphans
ExecStop=/usr/local/bin/docker-compose -f /opt/collector/collector.yaml down
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
```

