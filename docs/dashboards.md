# Dashboards

## Single Dashboard

To only show a single dashboard, simply add all configuration
and rules in the main `config.toml`.

### Example

`tuwat -conf config.toml`

```toml
[[rule]]
description = "Ignore Drafts"
[rule.label]
Draft = "true"

[[github]]
Repos = ['synyx/tuwat', 'synyx/buchungsstreber']
Tag = 'gh'
```

## Multiple Dashboards

To have multiple dashboards, add the main configuration to
the `config.toml` and create a folder containing more
rule files.

There are two kinds of dashboards:

* `mode = "exclude"`: The normal kind of dashboard.  Each rule
  will filter the matching items from the board.
* `mode = "include"`: Only items matching the rules are shown
  on the board.

### Example

`tuwat -conf config.toml -dashboards tuwat.d`

```toml
# config.example.toml
[[github]]
Repos = ['synyx/tuwat', 'synyx/buchungsstreber']
Tag = 'gh'
```

```toml
# tuwat.d/no-drafts.toml
[[rule]]
description = "Ignore Drafts"
[rule.label]
Draft = "true"
```

```toml
# tuwat.d/drafts.toml
[main]
mode = "include"

[[rule]]
description = "Show only drafts"
[rule.label]
Draft = "true"
```
