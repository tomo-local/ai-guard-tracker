# CommandGatePolicy Schema Reference

This is the full field reference for `.guard/policies/**/*.yaml` manifests
(the `CommandGatePolicy` resource). For a narrative introduction, see
[Writing Policies](../README.md#writing-policies) in the main README; this
document is the exhaustive spec, including every validation rule `guard
lint` enforces.

## File layout

- One file = one `CommandGatePolicy` resource = one tool + one `action`.
- Files are loaded recursively from the policies directory (`.guard/policies`
  by default, overridable with `GUARD_POLICIES_DIR`/`-rules`), so policies
  for the same tool are conventionally grouped under a subdirectory:
  `.guard/policies/<tool>/<short-action-or-qualifier>.yaml` (e.g.
  `git/allow.yaml`, `git/destructive-deny.yaml`, `rm/deny.yaml`).
- Both `.yaml` and `.yml` extensions are loaded; any other extension is
  ignored.

## Top-level structure

```yaml
apiVersion: ai-guard/v1       # required, must equal exactly "ai-guard/v1"
kind: CommandGatePolicy       # required, must equal exactly "CommandGatePolicy"
metadata:
  name: <string>              # required, non-empty, kebab-case convention
spec:
  target:
    module: permission        # required, must equal exactly "permission" in v1
    tool: <string>             # required, non-empty, e.g. "git", "rm", "curl"
  action: ALLOW | DENY         # required, exactly one of these two values
  rules:                       # required, at least 1 entry
    - <Rule>
    - <Rule>
    ...
```

`spec.action` applies to **every** rule in this resource -- a resource
can't mix ALLOW and DENY rules. If a tool needs both, split it across two
files (see [Multi-file example](#multi-file-example)).

## Rule

```yaml
- to:
    mode: read | write | destructive | network   # required
    functions: [sub, commands]                    # optional
    args:
      anyOf: ["-f", "--force"]                     # optional
  options:                                         # optional
    confirm:
      enabled: true | false
      message:
        reason: "..."
        next: "..."
    customCheck:
      path: "./scripts/x.sh"
      command: "..."
      args: ["..."]
      timeoutSeconds: 10
      when:
        - key: exit
          value: 0
          action: ALLOW | DENY
          confirm: true | false
  message:
    reason: "..."
    next: "..."
```

### `to` (match condition)

| Field | Type | Required | Notes |
|---|---|---|---|
| `to.mode` | string (enum) | yes | One of `read`, `write`, `destructive`, `network`. Categorization only -- doesn't affect which rule wins, just reporting/clarity. |
| `to.functions` | list of strings | no | Subcommand names this rule applies to (e.g. `[add, commit, push]`). **Omit entirely** to make the rule a catch-all for every invocation of the tool (used for flat commands like `rm`/`curl` that have no subcommand concept). |
| `to.args` | object | no | See below. |
| `to.args.anyOf` | list of strings | yes, if `args` is set | At least 1 non-empty string. The rule only matches commands whose argv contains **at least one** of these tokens, literally (no regex/globs, no flag-value parsing). |

How a command's subcommand is detected: the first argument that doesn't
start with `-` is treated as the "function". This is a heuristic -- it
misclassifies a flag's separate value as the function (e.g. `git -C <path>
push` sees `<path>` as the function, not `push`). See [Known
Limitations](../README.md#known-limitations-v1).

Matching precedence within one resource's `rules[]`: a rule naming
`functions` that lists the command's detected function wins; if none
matches, the first catch-all rule (no `functions`) applies. `args.anyOf`,
when set, further restricts either kind -- a rule with an `args` condition
that doesn't match is skipped entirely, even if its `functions`/catch-all
would otherwise apply.

### `options` (optional modifiers)

| Field | Type | Required | Notes |
|---|---|---|---|
| `options.confirm` | object | no | See [Confirm](#optionsconfirm). |
| `options.customCheck` | object | no | See [CustomCheck](#optionscustomcheck). |

#### `options.confirm`

| Field | Type | Required | Notes |
|---|---|---|---|
| `options.confirm.enabled` | bool | no (default `false`) | When `true` **and** the resource's `spec.action` is `ALLOW`, the effective decision becomes `CONFIRM` instead of a silent `ALLOW`. **`guard lint` rejects `enabled: true` when `spec.action` is `DENY`** -- DENY is meant to be unconditional, never negotiable via confirmation. |
| `options.confirm.message` | object | no | A `Message` (see below) specific to this confirmation gate. Takes precedence over the rule's own top-level `message` when both are set -- see [Message precedence](#message-precedence). |

#### `options.customCheck`

Mirrors Istio's `action: CUSTOM` -- delegates part of the decision to an
external script or inline shell command.

| Field | Type | Required | Notes |
|---|---|---|---|
| `options.customCheck.path` | string | exactly one of `path`/`command` | Path to an executable script, run directly with `args`. |
| `options.customCheck.command` | string | exactly one of `path`/`command` | Inline shell command, run via `sh -c "<command>" <args...>`. |
| `options.customCheck.args` | list of strings | no | Extra arguments passed to `path` or appended after `command`. |
| `options.customCheck.timeoutSeconds` | int | no (default `10`) | Must be `>= 0`. If the process doesn't exit in time, it's killed and treated as "no `when` match" (falls back to the rule's base action/confirm as usual). |
| `options.customCheck.when` | list of `When` | no | Evaluated in order against the process's exit code; the **first** match wins. If none match (including on timeout or a run error), the customCheck contributes no override. |

`When` entries:

| Field | Type | Required | Notes |
|---|---|---|---|
| `when[].key` | string | yes | Must equal exactly `"exit"` in v1 -- the only supported check key. |
| `when[].value` | int | yes | The exit code to match. |
| `when[].action` | string (enum) | yes | `ALLOW` or `DENY`. |
| `when[].confirm` | bool | no (default `false`) | When `true` **and** `when[].action` is `ALLOW`, the override resolves to `CONFIRM` instead of plain `ALLOW`. **`guard lint` rejects `confirm: true` when `when[].action` is `DENY`**, for the same reason as `options.confirm.enabled`. |

> **Security note**: `customCheck` scripts/commands run with the same
> trust boundary as the policy files themselves (committed to and
> reviewed in this repo). Never point it at an untrusted script.

### `message` (rule-level)

| Field | Type | Required | Notes |
|---|---|---|---|
| `message.reason` | string | no | Sent to the AI verbatim as Claude Code's `permissionDecisionReason` -- **write it in English**, not shown as end-user UI text. |
| `message.next` | string | no | Same as above; what the AI should do next (e.g. "ask the user before proceeding"). |

Either field may contain the literal placeholder `{cmd}`, replaced with
the exact matched command string (e.g. `git add -A .`). It's the only
placeholder supported in v1.

#### Message precedence

A rule may define its own top-level `message` (the general-purpose
reason/next for whatever this rule decides -- most useful for a DENY rule
explaining why it denies) **and/or** `options.confirm.message` (wording
specific to the confirmation gate). When both are set on the same rule,
**`options.confirm.message` wins**. When neither is set, no reason/next is
surfaced for that match.

## Evaluation semantics

This section documents how `guard check`/`internal/engine` actually
resolves a command against the loaded policy set -- useful when a rule
isn't behaving the way you expect.

1. **Decomposition**: the raw Bash command line is parsed with a real
   shell AST parser and walked for every `CallExpr` -- so `&&`, `;`, `|`,
   subshells, and command substitution (`$(...)`) are all decomposed into
   individual `<tool> <args...>` invocations and evaluated independently.
   Non-literal arguments (variable expansions, etc.) are rendered back as
   raw shell syntax, so they won't match a `functions` list.
2. **Per-command matching**: for each decomposed command, every policy
   resource targeting that `tool` is checked; within each, [`to`
   matching](#to-match-condition) picks at most one rule.
3. **Strictest-wins across matching resources**: if multiple resources'
   rules match the same command, the strictest `spec.action` among them
   wins (`DENY > CONFIRM > ALLOW`, with the confirm-derived `CONFIRM`
   ranking above plain `ALLOW`). Messages are deduplicated and joined
   across every matching resource/rule that tied at the strictest action.
4. **Confirm modifier**: if the strictest action among matches is `ALLOW`
   and any tied rule has `options.confirm.enabled: true`, the effective
   action becomes `CONFIRM`.
5. **CustomCheck override, if present**: when any tied rule has
   `options.customCheck` (the first one found is used), the script/command
   runs and its `when` list is evaluated against the exit code. A match
   **always overrides** steps 3-4's result -- even downgrading a
   confirm-gated `ALLOW` back to plain `ALLOW`. No match (including on
   timeout) leaves the result from steps 3-4 untouched. In short:
   **`customCheck` > `confirm` > the plain base action.**
6. **No match at all**: if no resource's rule matches a decomposed
   command, the built-in default applies -- action `CONFIRM`, a generic
   reason naming the unmatched command, and a generic "review and confirm"
   next step. This default is never writable as a YAML `action`.
7. **Aggregating a compound command**: steps 1-6 produce one `Decision`
   per decomposed command; the final decision for the whole command line
   is, again, strictest-wins across all of them (same ranking as step 3).

## `guard lint` rules

Beyond the field-level constraints in the tables above (required-ness,
enums, the path/command and `when[].key` checks), `guard lint` also
checks:

| Severity | Check |
|---|---|
| Error | `apiVersion` != `"ai-guard/v1"` |
| Error | `kind` != `"CommandGatePolicy"` |
| Error | `metadata.name` empty |
| Error | `spec.target.module` != `"permission"` |
| Error | `spec.target.tool` empty |
| Error | `spec.action` not one of `ALLOW`/`DENY` |
| Error | `spec.rules` empty |
| Error | `rules[].to.mode` not one of `read`/`write`/`destructive`/`network` |
| Error | `rules[].to.args.anyOf` set but empty, or containing an empty string |
| Error | `rules[].options.customCheck` has neither `path` nor `command`, or has both |
| Error | `rules[].options.customCheck.when[].key` != `"exit"` |
| Error | `rules[].options.confirm.enabled: true` together with `spec.action: DENY` |
| Error | `rules[].options.customCheck.when[].confirm: true` together with that entry's `action: DENY` |
| Warning | A second catch-all (no `functions`) rule in the same resource -- the first one already makes it unreachable |
| Warning | Two or more resources declare the exact same `{tool, mode, functions, args}` match condition -- not an error (strictest-wins resolves it deterministically), but almost always unintentional |

## Examples

### Minimal ALLOW with confirm

```yaml
# .guard/policies/curl/confirm.yaml
apiVersion: ai-guard/v1
kind: CommandGatePolicy
metadata:
  name: curl-confirm
spec:
  target:
    module: permission
    tool: curl
  action: ALLOW
  rules:
    - to:
        mode: network
      options:
        confirm:
          enabled: true
          message:
            reason: "{cmd} accesses an external network resource."
            next: "Confirm the URL is safe with the user before proceeding."
```

### Flag-level matching (DENY wins over CONFIRM)

```yaml
# .guard/policies/git/push-force-deny.yaml
apiVersion: ai-guard/v1
kind: CommandGatePolicy
metadata:
  name: git-push-force-deny
spec:
  target:
    module: permission
    tool: git
  action: DENY
  rules:
    - to:
        mode: destructive
        functions: [push]
        args:
          anyOf: ["-f", "--force"]
      message:
        reason: "{cmd} force-pushes and can overwrite remote history."
        next: "Suggest a non-destructive alternative, or ask the user to run it themselves."
```

Paired with a plain `git/allow.yaml` ALLOW+confirm rule for `push` with no
`args` condition: `git push` alone evaluates to `CONFIRM`, while `git push
-f`/`--force` evaluates to `DENY` (both rules match; `DENY` is stricter).

### CustomCheck with confirm fallback

```yaml
# .guard/policies/git/push-custom-check.yaml
apiVersion: ai-guard/v1
kind: CommandGatePolicy
metadata:
  name: git-push-custom-check
spec:
  target:
    module: permission
    tool: git
  action: DENY             # safe default: deny unless the script proves otherwise
  rules:
    - to:
        mode: write
        functions: [push]
      options:
        customCheck:
          path: "./scripts/check-push.sh"
          args: ["--staged"]
          timeoutSeconds: 10
          when:
            - key: exit
              value: 0
              action: ALLOW
              confirm: true   # still confirm with a human even when the script says it's safe
          # no `when` match (e.g. exit != 0, or timeout) falls back to spec.action = DENY
      message:
        reason: "{cmd} requires the push-safety script to pass before it runs."
        next: "Review the script's output with the user before proceeding."
```

### Multi-file example

One tool, two actions, two files:

```yaml
# .guard/policies/git/allow.yaml
spec:
  action: ALLOW
  rules:
    - to: { mode: read, functions: [log, status, diff] }
    - to: { mode: write, functions: [add, commit, push] }
      options:
        confirm: { enabled: true }
```

```yaml
# .guard/policies/git/destructive-deny.yaml
spec:
  action: DENY
  rules:
    - to: { mode: destructive, functions: [reset, clean, rebase] }
      message:
        reason: "{cmd} is a destructive operation and is denied."
        next: "Suggest an alternative approach to the user."
```
