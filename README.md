# Tuwat Dashboard

## What is the tuwat Dashboard?

Tuwat is a replacement for the venerable [Nagdash] which had been adapted
time and time again to provide accessors for an evolving infrastructure.

The driving mindset for Tuwat (in German "tu was", meaning "do something")
is to show actionable items.  This is a  slight departure of [Nagdash],
which shows only shows Nagios `Hosts`/`Services`.

[Nagdash]: https://github.com/lozzd/Nagdash

## Features

Connectors for

* Prometheus [Alertmanager]
* [GitLab] MRs
* [Icinga 2]
* [Nagios API]
* [Patchman]

[Alertmanager]: https://prometheus.io/docs/alerting/latest/alertmanager/
[GitLab]: https://www.gitlab.com
[Icinga 2]: https://icinga.com
[Nagios API]: https://github.com/zorkian/nagios-api
[Patchman]: https://github.com/furlongm/patchman

## Development

```shell
go build -o tuwat ./cmd/tuwat
./tuwat -conf msis.toml -environment test -mode dev
```
