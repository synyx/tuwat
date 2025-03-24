# Releases

## 1.13.0 - 2025-03-24 Alert Groups

* Alerts can now be grouped via `group_alerts = true`
  This only affects the view.
* Hosts in DOWN state will now be marked as CRITICAL for Icinga2/Nagios.

## 1.12.3 - 2025-01-03 New Year

* Maintenance release, updating dependencies.

## 1.12.1 - 2024-10-17 Patchman Slowness

* Handle slow patchman instances.

## 1.12.0 - 2024-10-17 Patchman Filter

* Enable pre-filtering Patchman results, see [example.toml](config.example.toml)
  for details.

## 1.11.0 - 2024-10-16 Menu Styling

* The menu has been given an overhaul to be more readable in the dark theme.
* The current dashboard is now better recognizable in the menu.

## 1.10.2 - 2024-09-03 Log Volume Fix

* The 1.10.0 version unfortunately introduced a large unnecessary log volume
  change.

## 1.10.1 - 2024-08-29 Graylog Labels

* Add Graylog event grouping fields to labels.

## 1.10.0 - 2024-08-29 Graylog

* Add Graylog connector.

## 1.9.1 - 2024-08-09 Maintenance

* Only dependency upgrades

## 1.9.0 - 2024-07-10 Tracing IDs

* Using `tuwat -otelUrl stdout` or `TUWAT_OTEL_URL=stdout` now enables tracing
  output on stdout.
* Add spans/traces to logs to enable correlation
* Cleanup file paths in logging for visibility by stripping the build path.
* Remove error logging keys containing `<nil>` strings.

## 1.8.0 - 2024-06-05 Icinga2 Host ACKs

* Downtimes and acknowledgements as well as disabling notifications for a host
  will also disable those for the services on this host. This differs from
  `icinga` and `nagios`, as they allow actions like
  `Schedule downtime for this host` vs. `allow downtime for
  this host and all services`.
* Potentially breaking: Change logging system from `zap` to stdlib
  `slog`. In case the logs are parsed using external systems, the logs are now
  in a different format. This was done to reduce the overall amount of
  dependencies.

## 1.7.0 - 2024-04-29 Negative Times

* Allows times to be in the future (negative seconds ago)
  and thus to be able to filter those correctly.

## 1.6.0 - 2024-02-21 GitLab MR groups

* Allow GitLab Groups to be configured for pulling merge requests. This has the
  advantage of needing to specify fewer single projects and also ignores
  archived projects by default.
* Fix application crash on certain network timeouts.

## 1.5.2 - 2024-02-20

* Bugfix release for `when` rules only being applied if no `where`
  rule given.

## 1.5.1 - 2024-02-19

* Bugfix release for `when` rules and `~=` matching rule prefixes.

## 1.5.0 - 2024-02-19 backend mode

* Add `alertmanager` API, e.g. `/api/alertmanager/v2/status` for use as API
  project. Note that this is an experimental feature and should not be relied
  upon.
* Add negative matchers: `!=` and `!~`
* Use [semver](https://semver.org/) for versioning tuwat. This allows tuwat to
  be used as a library.

## v1.4 - 2024-01-11 light mode

* Make the look of tuwat configurable via `style` property.

## v1.3 - 2024-01-08 icinga2 ack

* The `icinga2` adapter now filters all acknowledged issues.

## v1.2 - 2023-12-19 Stability

* Fix rule matching from `v1.1` which made every rule match everything.
* `tuwat` now listens to every network instead of only `localhost`
  for the web port. The management port still only binds to `localhost`.
* The `icinga2` adapter now adds host groups to services.
* Add more documentation regarding rules and dashboards.

## v1.1 - 2023-11-20 Rule Matchers

* Add a `when` rule to be able to match on when the alert happened.
* Add more expressive rules, making it possible to express a
  `greater than` on numerical values. See `README` for details.
* Breaking configuration changes:
    * The `what` rules are now combined via `AND` with label rules. This
      streamlines the behaviour, making it behave like the label rules
      themselves. It also makes it possible e.g. to express that a rule matcher
      only applies when the alert is old.

## v1.0 - 2023-10-23 Maintenance

* Revise look of alerts
* Add example connector

## v0.19 - 2023-09-12 Visuals

* Make red alert more readable
* Make filtering Redmine alerts more functional

## v0.18 - 2023-08-02 Redmine

* Add Redmine connector
* Add `orderview` connector
* Recognize service downtimes for `nagios-api` connector
* The filtered alert-list is now hidden by default and can be toggled

## v0.17 - 2023-06-26 Multiple Dashboards

* UX change: Full row for items now clickable for details
* Adds possibility of having multiple dashboards/rule-sets
* Breaking configuration changes:
    * OAuth2 configuration is now a subsection named `OAuth2Creds`
    * BasicAuth Username now configured with key `Username` instead of `User`
      for all connectors.
* Breaking behavioural changes:
    * If observability depends on the label `thing` for websocket connections,
      this has been renamed to `client`.

## v0.16 - 2023-06-12 Resource Consumption

* To not display all merge requests, you can now specify the projects via
  `Projects = []`
* Fixes connection leak in nagios-api connector
* Breaking Configuration Changes:
    * Change parsing of rules, so that label matchers are combined with `AND`
      instead of `OR`.

## v0.15 - 2023-05-26 GitLab MR flood

* Handle more open merge requests from GitLab

## v0.14 - 2023-05-16 Stability

* Make management port configurable via `-mgmtAddr :8987`
* Add `net/pprof` for debugging, see http://127.0.0.1:8987/debug/pprof
* Add `Down` state for health endpoint if last collection too old

Breaking Changes:

* actuator endpoints now on different port, no longer on main port (`8988` by
  default)
    * http://127.0.0.1:8987/actuator/health (`/info`, `/prometheus`)

## v0.13 - 2023-05-11 Multiple Users

* Fix websocket/sse registrations
* Make event source selectable via /alerts?eventSource=sse|websocket|fallback

## v0.12 - 2023-05-10 SSE

* Use SSE if Websockets are not available

## v0.11 - 2023-05-06 GitHub Collector

* Add a new GitHub collector showing open issues/pull requests
* Breaking Configuration Changes:
    * `GND_ENVIRONMENT` env var renamed to `TUWAT_ENVIRONMENT`
    * `GND_INSTANCE` env var renamed to `TUWAT_INSTANCE`
    * `GND_ADDR` env var renamed to `TUWAT_ADDR`
    * `-mode dev|prod` removed in favour of `TUWAT_TEMPLATEDIR` and
      `TUWAT_STATICDIR` env vars.
* Added `TUWAT_CONF` env var for specifying the configuration file

## v0.10 - 2023-02-10 Minimize Collections

* Do not collect, if there is no-one subscribed to the dashboard to avoid
  cpu/network activity
* Propagate telemetry information to sources

## v0.9 - 2022-10-31 Open Source

* Open the source under a 3-Clause BSD License
* Preparations for releasing on GitHub
* Rename project from a temporary `gonagdash` to a hopefully less temporary
  `tuwat`.
* Revise output on `tuwat -version`

## v0.8 - 2022-10-24 Silencing

* Add preview functionality to silence alerts

## v0.7 - 2022-10-03 UI / GitLab

* Revise link symbols
* Revise use of GitLab connector, now targeting all visible MRs

## v0.6 - 2022-09-28 Filter Configuration

* Revise how filters are configured

## v0.5 - 2022-09-27 Stability

* More tags for filtering Patchman
* Make collection interval configurable
* Add `SIGHUP` for reloading the configuration

## v0.4 - 2022-09-27 Styling

Add links for some connectors and revise the UI to display more information.

## v0.3 - 2022-09-25 Push Revisions

Bugfixes all over.

## v0.2 - 2022-09-25 Server Side Push

Stream changes to the browser via Websockets.

## v0.1 - 2022-09-24 Initial Release

Initial Release of Tuwat providing a basic non-working Dashboard for Nagios,
Icinga 2, Patchman, Alertmanager and GitLab MRs.
