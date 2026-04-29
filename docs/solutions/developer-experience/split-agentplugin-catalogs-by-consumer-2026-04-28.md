---
title: Split AgentPlugin Catalogs by Consumer
date: 2026-04-28
category: developer-experience
module: ATV plugin generation
problem_type: developer_experience
component: tooling
severity: medium
applies_when:
  - "A source-installed AgentPlugin repository also exposes a granular CLI marketplace"
  - "VS Code install UX should show one product option while CLI users keep advanced packages"
  - "A generator drift check runs on Windows checkouts with Git autocrlf enabled"
tags: [vscode, agentplugin, marketplace, plugingen, autocrlf]
---

# Split AgentPlugin Catalogs by Consumer

## Context

ATV needed one clean VS Code `Chat: Install Plugin from source` option without removing its advanced Copilot CLI plugin catalog. The confusing state was that VS Code could fall through to the granular `.github/plugin/marketplace.json` catalog and show dozens of packs and single-skill plugins to a first-time user.

The repo also needed generator validation to work on Windows. Git may check generated JSON and Markdown out with CRLF line endings even though `plugingen` emits LF, so a byte-only drift check can report false positives on an otherwise in-sync tree.

## Guidance

Use separate generated catalog surfaces for separate consumers:

- Root `marketplace.json`: curated VS Code source-install catalog. It should expose exactly one product entry, `atv-starter-kit`, with `source` pointing to `./plugins/atv-everything`.
- `.claude-plugin/marketplace.json`: byte-identical mirror of the root source-install catalog for Claude-format compatibility.
- `.github/plugin/marketplace.json`: Copilot CLI catalog. It can keep every granular package, but should put the full `atv-everything` bundle first.

Keep the generator as the source of truth. Do not hand-maintain the catalogs independently. `Generate` should write all three surfaces, and `CheckClean` should compare them all.

When checking generated output, normalize line endings during comparison:

```go
func equalGeneratedContent(a, b []byte) bool {
    if bytes.Equal(a, b) {
        return true
    }
    return normalizeLineEndings(string(a)) == normalizeLineEndings(string(b))
}
```

Also prefix drift paths by output root before reporting them. Without prefixes, failures from root `marketplace.json`, `.claude-plugin/marketplace.json`, and `.github/plugin/marketplace.json` can all appear as repeated `marketplace.json`, which slows review.

## Why This Matters

The same repository can serve different install surfaces, but the UX contract is different for each one. VS Code source install is a product-selection moment, so it should present one coherent ATV option. Copilot CLI users may still need granular packs, single-skill installs, or agents-only installs, so the CLI catalog can remain broad.

Separating the catalogs lets ATV improve the VS Code first-run experience without breaking CLI compatibility. Keeping all catalogs generated prevents drift between the product decision and the checked-in metadata.

Line-ending tolerant comparisons make the generator check portable. The check still catches real content drift, missing files, stale files, and missing trailing newlines, but it does not fail just because the working tree is CRLF on Windows.

## When to Apply

- A repo has one internal/plugin packaging graph but multiple consumer-facing installers.
- VS Code or another source-install flow reads catalog files in precedence order and can accidentally pick up an expert-oriented catalog.
- Generated docs or metadata are checked on both Unix-like and Windows developer machines.

## Examples

The intended catalog invariant is:

```text
marketplace.json                         -> 1 entry: atv-starter-kit -> ./plugins/atv-everything
.claude-plugin/marketplace.json          -> byte-identical mirror of root marketplace.json
.github/plugin/marketplace.json          -> 42 CLI entries, atv-everything first
plugins/atv-everything/.claude-plugin/   -> per-plugin source-install metadata
```

The regression tests should make those invariants explicit:

```go
if len(mp.Plugins) != 1 {
    t.Fatalf("source-install marketplace should expose exactly one plugin, got %d", len(mp.Plugins))
}
if entry.Source != "./plugins/atv-everything" {
    t.Errorf("source-install entry source: got %q want ./plugins/atv-everything", entry.Source)
}
```

And the CLI test should continue to protect the granular catalog:

```go
if len(mp.Plugins) != len(wantNames) {
    t.Errorf("marketplace.json plugin count: got %d want %d", len(mp.Plugins), len(wantNames))
}
if len(mp.Plugins) == 0 || mp.Plugins[0].Name != "atv-everything" {
    t.Fatalf("CLI marketplace should put atv-everything first, got %+v", mp.Plugins)
}
```

## Related

- `pkg/plugingen/generate.go`
- `pkg/plugingen/generate_test.go`
- `docs/marketplace.md`
- `docs/plans/2026-04-28-001-feat-vscode-source-install-agentplugin-update-plan.md`
