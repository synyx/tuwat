# Rules

The rule-system works via an exclude list, matching rules will include or
exclude the matching items, depending on the
[type of dashboard](dashboards.md#dashboard-types).

The configuration is done via [toml](https://toml.io/).

For example:

```toml
[[rule]]
description = "blocked because not needed"
what = "fooo service"
```

* The `description` field provides a visible explanation, why the item is
  excluded.
* The `what` field selects all items where the `What` matches the given
  regular expression.

```toml
[[rule]]
description = "Ignore Drafts"
what = "Thing"
when = "> 60"
[rule.label]
Draft = "true"
```

* The `label` section selects items via labels.  In this example it would match
  an item which has the label `Draft` which matches the given regular expression.
* The label rules will combine as `AND` with other label rules and `when` and
  `what` rules.
* `when` rules interpreted as "X seconds from now".  The above example would match
  an alert when the alert has lasted a minimum of 60 seconds.  Times in the future
  have an undefined behaviour.

## Matching Rules

The default is to match the value in the configuration as a regular expression.
However, this can be changed by specifying an operator.

* `~= string`: Explicitly require a regular expression to be matched.
  This is the same as just leaving `~= ` out.
* `=  string|number`: Require the string or number to exactly match.  In case  
  the value is numeric, this will mean that the value will compared like a
  floating point value.  This means that differences below `1e-8` will be
  considered to be the same.
* `>  number`: Require both configuration and the value in the alert to be a
  numerical value and that the value in the alert to be bigger than the
  configured number.
  This also applies to the `<`, `>=`, `<=` operators.

### Example

```toml
[[rule]]
description = "Ignore certain group"
[rule.label]
groups = "= A-Group"

[[rule]]
description = "Ignore all with group beginning with A"
[rule.label]
groups = "~= (^|,)A"

[[rule]]
description = "Ignore old things"
when = ">= 60"
```
