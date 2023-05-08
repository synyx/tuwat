# Tuwat Dashboard

## What is the tuwat Dashboard?

Tuwat is a replacement for the venerable [Nagdash] which had been adapted
time and time again to provide accessors for an evolving infrastructure.

The driving mindset for Tuwat (in German "tu was", meaning "do something")
is to show actionable items.  This is a  slight departure of [Nagdash],
which shows only Nagios `Hosts`/`Services`.

[Nagdash]: https://github.com/lozzd/Nagdash

## Features

Connectors for

* Prometheus [Alertmanager]
* [GitLab] MRs
* [GitHub] PRs
* [Icinga 2]
* [Nagios API]
* [Patchman]

[Alertmanager]: https://prometheus.io/docs/alerting/latest/alertmanager/
[GitLab]: https://www.gitlab.com
[GitHub]: https://www.github.com
[Icinga 2]: https://icinga.com
[Nagios API]: https://github.com/zorkian/nagios-api
[Patchman]: https://github.com/furlongm/patchman

## License

[BSD 3-Clause License](./LICENSE)

## Development

### Local Development

```shell
go build -o tuwat ./cmd/tuwat
export TUWAT_TEMPLATEDIR= TUWAT_STATICDIR=
./tuwat -conf config.example.toml
```

* Open http://localhost:8988

### JavaScript Development

Updating the `main.js` used by the HTML code:

* Update JavaScript dependencies in `package.json`/`package-lock.json`
* Edit code in `pkg/web/static/js/index.js`

```shell
npm run build # to generate the bundled files
npm run watch # to watch for changes and re-generate while developing
```

Make sure to add the changed/generated files, so not everyone has to use
nodejs.
