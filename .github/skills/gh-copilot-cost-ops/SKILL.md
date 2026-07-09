---
name: gh-copilot-cost-ops
description: >
  Build real cost-ops infrastructure for GitHub Copilot usage-based billing:
  per-PR token cost attribution, cost cliff detection, semantic dedup layers,
  token-aware agent wrappers, and programmatic budget control. Use when
  engineering teams need visibility and control over AI credit spend — not
  checklists, but running software. Triggers on 'copilot cost', 'ai credits',
  'token spend', 'ubb tooling', 'gh copilot billing infra', or when building
  cost-aware agent infrastructure.
argument-hint: "[what to build or diagnose]"
---

# gh-copilot-cost-ops

GitHub Copilot shifted to usage-based billing June 1, 2026. The public docs,
blog posts, and sales enablement material are 101 — buyers have already read
them. This skill is for engineering teams who need to build the infrastructure
that does not exist yet.

**Research reference:** `docs/context/research/2026-07-09-github-copilot-ubb-tbb-deep-dive.md`

## What this is NOT

- Not a cost calculator or estimator form
- Not a "tips for tight prompting" checklist
- Not a redo of the GitHub docs

## The 0→1 Problem Space

The core gap: **zero feedback loop** between work happening and money burning.
Usage appears in a monthly billing report 2–4 weeks after the session ended.
By then the behavior that caused it is long gone.

Five engineering problems worth solving:

| Problem | Why it matters |
|---------|---------------|
| No per-session cost signal | Agentic sessions can cost 50× more than chat; developers are flying blind |
| No per-PR/workflow attribution | Can't tell which codebase, team, or workflow pattern is driving spend |
| No cost-aware agent behavior | Agents don't know their budget; they run until blocked or done |
| Semantic deduplication | Same prompt answered 47× in a week; each one bills full price |
| Programmatic budget control | Budget hierarchy exists but no IaC, no API wrappers, only UI |

## GitHub APIs and Data Sources

**Billing API (requires `manage_billing:copilot` scope):**
```
GET /enterprises/{enterprise}/copilot/billing/seats
GET /enterprises/{enterprise}/copilot/usage
GET /orgs/{org}/copilot/usage
```
Usage response includes: `day`, `total_engaged_users`, `total_ide_code_completions`,
`total_ide_chat`, `total_github_dotcom_chat`, `total_pr_summaries` — broken down
by model and editor. No per-session granularity yet.

**Billing preview report (CSV download, admin only):**
Columns: user, model, surface, input tokens, output tokens, cached tokens,
AI credits. Available from `github.com/[org]/settings/billing`.

**Budget controls API (as of June 2026):**
```
POST /enterprises/{enterprise}/settings/billing/usage_report_downloads
GET /enterprises/{enterprise}/settings/billing/spending_limit
PATCH /enterprises/{enterprise}/settings/billing/spending_limit
```

**Key constraint:** No per-session or per-PR token data in the API today.
Attribution to a PR requires correlating timestamps + user + model with
PR activity — a join, not a direct query.

## Architecture Patterns

### 1. Per-PR Cost Attribution (CI GitHub Action)

```
PR opens / updates
  → GitHub Action triggers
  → Record: pr_number, sha, author, timestamp_start
  → Wait for Copilot usage delta (poll billing API with T+5min window)
  → Post comment: "AI credits consumed on this PR: ~{X} credits (~${Y})"
```

**Challenge:** Billing API granularity is daily, not per-event. Best approach:
baseline the org's daily usage before the PR window, diff after, attribute
the delta to active PRs proportionally. Imprecise but directional — exactly
what the GitHub field guidance says about their own reports.

**Proven pattern:** Store baseline in Action cache keyed by date, diff at
job end. Works within GitHub Actions limits.

### 2. Cost Cliff Detector (MCP Server / VS Code Extension)

Real-time running cost display during agent sessions.

```
Local MCP server
  → Intercepts Copilot API calls (input/output token counts)
  → Maintains session running total
  → Emits: current spend, projected session cost, warning at threshold
  → Exposes: /cost/current, /cost/session, /cost/warn (SSE stream)
```

Token counts are available in VS Code extension telemetry events.
The Copilot VS Code extension exposes `vscode.env.machineId`-scoped usage.

**Simpler path:** Read from VS Code's `github.copilot.usage` telemetry
output file (`.copilot/usage.jsonl` if enabled) rather than intercepting.

### 3. Semantic Dedup Cache (MCP Tool)

Catch semantically identical prompts before they hit the model.

```
Prompt arrives
  → Embed with lightweight model (text-embedding-3-small, ~$0.02/1M tokens)
  → Query local vector store (SQLite + sqlite-vec, or Chroma)
  → If cosine similarity > 0.92: return cached response
  → Else: forward to model, cache response + embedding
```

**ROI math:** If 10% of prompts are near-duplicates, and average prompt costs
$0.05 in AI credits, 1000 prompts/day → saves $5/day → $150/month per heavy user.

**Stack:** SQLite + sqlite-vec (zero infra), or Chroma for team-shared cache.
Embedding model cost is negligible vs frontier model cost.

### 4. Cost-Aware Agent Wrapper

Agent that enforces a budget before starting and halts mid-run if exceeded.

```typescript
// Pseudocode
const session = new CostAwareSession({
  budget_credits: 500,          // $5 hard cap
  warn_at_credits: 400,         // warn at 80%
  on_warn: async () => ask_user("Continue? Current spend: {X} credits"),
  on_limit: () => { throw new BudgetExceededError() }
});

session.on('token_usage', ({ input, output, model }) => {
  const cost = calcCredits(input, output, model);
  session.accrue(cost);
});
```

**Integration point:** Wrap the Vercel AI SDK's `streamText`/`generateText`
callbacks, or hook into LangChain's `LLMResult` callbacks. Credit math lives
in a shared `@your-org/copilot-cost` package.

**Model credit rates:** See research note for full pricing table. Key rates
(per 1M tokens): GPT-5.4 input $2.50/$0.25 cached / $15.00 output;
Claude Sonnet 4.6 input $3.00/$0.30 cached / $3.75 write / $15.00 output.

### 5. Programmatic Budget Management

IaC-style wrapper over the GitHub budget controls API hierarchy.

```yaml
# copilot-budgets.yaml
enterprise: acme-corp
universal_user_budget: 20.00   # $20/user/month hard cap
cost_centers:
  - name: engineering
    user_budget: 50.00          # engineers get $50
    spending_limit: 5000.00     # team metered cap after pool
  - name: marketing
    user_budget: 5.00
    spending_limit: 500.00
overrides:
  - user: poweruser@acme.com
    budget: 150.00              # individual override
```

Apply with: `gh-cost-budget apply copilot-budgets.yaml`

**API calls needed:**
- `PATCH /enterprises/{e}/settings/billing/spending_limit` (enterprise level)
- Cost center APIs are newer — check `docs.github.com/copilot/concepts/billing/budgets-for-usage-based-billing` for current endpoints

## What to Build First

Prioritized by ROI-to-effort ratio:

1. **Semantic dedup MCP tool** — highest ROI, fully self-contained, no GitHub API auth needed
2. **Cost cliff detector MCP server** — visible in every session, immediate feedback
3. **Per-PR CI cost attribution** — GitHub Action, team-visible, drives behavior change
4. **Programmatic budget management** — `gh` CLI extension or standalone tool
5. **Cost-aware agent wrapper** — library, needs adoption

## Non-Goals

- Not a dashboard replacement for `github.com/settings/billing`
- Not an exact billing predictor (GitHub's own reports are directional)
- Not a usage reduction tool through model throttling — use GitHub's Auto routing
- Not a competitor to GitHub's native budget controls — complement them

## Stale When

- GitHub adds per-session token API (removes the attribution estimation need)
- Billing API gains per-PR granularity
- VS Code Copilot extension exposes a public cost telemetry API
- Model pricing table changes (check `docs.github.com/copilot/reference/copilot-billing/models-and-pricing`)

## Key References

- Research note: `docs/context/research/2026-07-09-github-copilot-ubb-tbb-deep-dive.md`
- Plan: `docs/plans/2026-07-09-000-kb-ubb-tbb-tooling-manifest.md`
- GitHub billing API: `https://docs.github.com/en/rest/copilot/copilot-usage`
- Budget concepts: `https://docs.github.com/en/copilot/concepts/billing/budgets-for-usage-based-billing`
- Model pricing: `https://docs.github.com/en/copilot/reference/copilot-billing/models-and-pricing`
