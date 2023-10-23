# Releases

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
* Add net/pprof for debugging, see http://127.0.0.1:8987/debug/pprof
* Add Down state for health endpoint if last collection too old

Breaking Changes:

* actuator endpoints now on different port, no longer on main port (`8988` by default)
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

* Do not collect, if there is no-one subscribed to the dashboard to
  avoid cpu/network activity
* Propagate telemetry information to sources

## v0.9 - 2022-10-31 Open Source

* Open the source under a 3-Clause BSD License
* Preparations for releasing on GitHub
* Rename project from a temporary `gonagdash` to a hopefully less
  temporary `tuwat`.
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

Initial Release of Tuwat providing a basic non-working Dashboard for
Nagios, Icinga 2, Patchman, Alertmanager and GitLab MRs.
