# Tuwat Dashboard

## What is the tuwat Dashboard?

Tuwat is a replacement for the venerable [Nagdash] which had been adapted time
and time again to provide accessors for an evolving infrastructure.

The driving mindset for Tuwat (in German "tu was", meaning "do something")
is to show actionable items. This is a slight departure of [Nagdash], which
shows only Nagios `Hosts`/`Services`.

[Nagdash]: https://github.com/lozzd/Nagdash

## Features

Connectors for

* Prometheus [Alertmanager]
* [GitLab] MRs
* [GitHub] PRs
* [Graylog] Events
* [Icinga 2]
* [Nagios API]
* [Patchman]
* Past due [Redmine] tickets
* Static example showing alert types
* [wiz.io] Issues

[Alertmanager]: https://prometheus.io/docs/alerting/latest/alertmanager/
[GitLab]: https://www.gitlab.com
[GitHub]: https://www.github.com
[Graylog]: https://graylog.org/
[Icinga 2]: https://icinga.com
[Nagios API]: https://github.com/zorkian/nagios-api
[Patchman]: https://github.com/furlongm/patchman
[Redmine]: https://redmine.org/
[wiz.io]: https://www.wiz.io/

## Configuration

See the [Example Config](config.example.toml) for configuration.

Available styles:

* `dark` (default)
* `light` - mimics the venerable nagdash

### Dashboards

The main configuration can contain `Rules`, but if multiple rule-sets/dashboards
are needed, dashboards can be added to a folder.

The `-dashboards` flag can be used to specify the folder, by default it looks at
`/etc/tuwat.d`.

The files have to end with `.toml`, the basename will be used as dashboard name.

For further examples and more information on dashboards, see the
[dashboard documentation](docs/dashboards.md).

### Rules

The rule-system works via an exclude list, matching rules simply exclude items.

For example:

```toml
[[rule]]
description = "blocked because not needed"
what = "fooo service"
```

For more information, see the [rule documentation](docs/rules.md).

## License

[BSD 3-Clause License](LICENSE)

## Development

### Local Development

```shell
go build -o tuwat ./cmd/tuwat
export TUWAT_TEMPLATEDIR= TUWAT_STATICDIR=
./tuwat -conf config.example.toml
```

* Open http://localhost:8988

Setting `TUWAT_TEMPLATEDIR` and `TUWAT_STATICDIR` to empty will automatically
use the development directories (`pkg/web/templates` and `pkg/web/static`
respectively). Not declaring the template/static directory means that the
versions bundled into the binary are used.

### Adding a new collector

* See `pkg/connectors/example` for a very basic example on how a connector is
  implemented.

### JavaScript Development

Updating the `main.js` used by the HTML code:

* Update JavaScript dependencies in `package.json`/`package-lock.json`
* Edit code in `pkg/web/static/js/index.js`

```shell
npm run build # to generate the bundled files
npm run watch # to watch for changes and re-generate while developing
```

Make sure to add the changed/generated files, so not everyone has to use
`node.js`.
