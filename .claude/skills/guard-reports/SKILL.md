---
name: guard-reports
description: Summarize ai-guard-tracker's check/record logs for the user and generate a shareable HTML report via `guard report`. Use when the user asks for a guard report, wants to see guard activity/stats, or runs /guard-reports.
---

# guard-reports

> **Status check first**: `guard report` is not implemented yet as of this
> writing (see `tmp/remaining-plan.md` in this repo if present). Before
> doing anything else, run `guard report -h` (or `go run . report -h`). If
> it fails with "unknown command" or similar, tell the user plainly that
> `guard report` hasn't shipped yet and stop here -- do not improvise a
> summary by hand-parsing `.guard/logs/*.jsonl` yourself; that duplicates
> work the real command is meant to do correctly (e.g. the strictest-wins
> and customCheck accounting) and will likely be misleading.

Once `guard report` exists, this skill's job is:

1. Run `guard report` with no flags and relay its text summary to the user
   in chat -- action/tool/mode counts, which tools are triggering
   deny/confirm most often, and anything about customCheck executions
   (exit codes, timeouts, which `when` overrides fired).
2. Generate the shareable artifact:
   ```bash
   guard report --format html --out .guard/report.html
   ```
   and tell the user the path (`.guard/report.html`). Don't print the raw
   HTML into chat.
3. If the logs are empty or very small, say so plainly rather than treating
   a thin report as if it were meaningful.
4. If the user asks for a specific slice (a date range, one tool, only
   denies), check `guard report -h` for supported flags before assuming
   none exist, and otherwise say that filtering isn't supported yet.
