---
title: Capability Modernization Workflow
description: Evidence-based replacement of customer-specific infrastructure with portable capabilities
---

## Start with the capability

Describe what the application needs without naming the current implementation. Derive functional and non-functional requirements from repository evidence before researching products.

Classify modernization as required for portability or optional tuning. Keep a working implementation until its replacement, compatibility layer, and rollback path are validated.

## Compare viable options

Compare two to four options, including retaining the current implementation behind an adapter when appropriate. Evaluate:

* Protocol and standards compatibility
* Portability and provider independence
* Authentication, authorization, private networking, and data residency
* Reliability, latency, scaling, failover, retry, and circuit breaking
* Policy enforcement, quotas, rate limits, budgets, and governance
* Observability, OpenTelemetry support, auditability, and cost telemetry
* Operational ownership, maturity, licensing, support, and lock-in
* Migration effort, testability, backward compatibility, and rollback

Research current official documentation when capabilities, support, pricing, availability, or licensing may have changed. Record sources and research date. Do not infer current product behavior from memory.

## Design application-owned boundaries

Prefer the repository's existing interface or introduce a narrow application-owned capability boundary. Keep provider authentication, endpoints, SDK configuration, request translation, and provider-specific errors inside adapters.

Use an anti-corruption layer when the current customer service exposes concepts that should not become the application's domain model. Add the replacement alongside the current provider before retiring production capability.

Provide a supported local-development path for customer-hosted capabilities. Prefer a contract-faithful mock, emulator, local adapter, or explicitly selected alternate provider. Do not bypass readiness, authentication, authorization, policy, or safety checks with an undocumented environment variable merely to make the application start. If a development bypass is approved, constrain it to a development environment, reject it in production configuration, emit a visible warning, document it, and test both the guarded and normal paths.

## Preserve the credential lifecycle

Debranding identity configuration must preserve how credentials are acquired,
stored, selected, refreshed, expired, and reloaded. Record the source precedence
for environment variables, mounted files, workload identity, and local credential
stores. Fail clearly when sources conflict instead of silently selecting whichever
value happens to be present.

Keep runtime tokens out of tracked environment files and templates. Use an ignored,
least-privilege runtime file or operating-system credential store for local refresh,
and deployment-native secret injection for automation. Maintain one writable runtime
source of truth rather than copying the same token into both an environment file and a
token file. Write files atomically with restrictive permissions. Never print, pass on a
command line, copy into a plan, or persist the token merely to make a readiness check pass.

Use the provider's supported authentication library and flow. Prefer authorization code
with PKCE, device code, managed identity, workload identity, or another flow appropriate
to the client. Interactive flows must use registered redirect URIs, cryptographically
unpredictable state and nonce values, and issuer, audience, signature, and expiry
validation. Do not ask users to paste bearer tokens through a public token-inspection
service or treat a browser opening as proof that authentication works. Test auth URL or
request construction with explicit fictional configuration and assert that missing
identity configuration fails before navigation.

When the documented workflow promises refresh without restart, validate it as a
behavioral contract: start the process or container, replace the credential through
the supported refresh path, and prove the next authenticated request observes the new
value without exposing either value. Mount only the runtime credential path needed for
reload rather than a broader source tree when the platform permits it.

Readiness checks should distinguish missing, placeholder, malformed, expired, and
unreachable authentication state before the first capability request. Decoding an
unverified token to inspect expiry is advisory readiness evidence only; it is never
proof of signature, issuer, audience, or authorization validity.

Select provider authentication from explicit configuration or an exact, parsed,
allowlisted destination. Do not use substring or suffix guesses that can leak a bearer
token to an attacker-controlled host. Test every supported provider route and a negative
cross-provider case proving each credential is withheld from all other destinations.
When the destination-specific credential is absent, malformed, or expired, fail closed;
do not fall back to a credential issued for another trust boundary.

## Assess LLM gateways

For an internal LLM gateway, evaluate:

* OpenAI-compatible or provider-neutral request contracts
* Multi-provider and multi-model routing
* Managed or workload identity and short-lived credentials
* Policy, quota, rate-limit, and budget controls
* Content safety and prompt-injection defenses
* Retry, fallback, circuit breaking, and regional failover
* Semantic caching when workload semantics support it
* OpenTelemetry traces, token usage, latency, and cost telemetry
* Model catalog, versioning, evaluation, and deployment governance
* Private networking, compliance, and data-residency requirements

Consider managed AI gateways, standards-compatible proxies, direct provider adapters, and open-source self-hosted gateways as solution classes. Do not predetermine the answer from the current cloud or customer name.

## Stage the migration

1. Characterize the current contract and dependent workflows.
2. Add contract tests or representative mocks around the capability boundary.
3. Place the current implementation behind an adapter without changing behavior.
4. Add the chosen provider alongside the current provider.
5. Validate configuration, authentication context, response semantics, telemetry, failure handling, and cost controls.
6. Roll out with a feature flag, routing policy, or environment-specific selection.
7. Define rollback conditions and preserve the previous adapter until acceptance criteria pass.
8. Remove customer-specific code only after dependent workflows no longer require it.

Validate the local path from a clean checkout without access to customer infrastructure. Exercise startup, health, and one representative capability request; a server that starts only after skipping the capability check is not portable.

When credentials are short-lived, repeat that representative request after refresh or
rotation without restarting the process when hot reload is part of the supported path.
Exercise missing, placeholder, malformed, and expired states, and verify that logs and
errors remain redacted.

Treat identity, certificates, network trust, externally owned workflows, provider replacement, immutable resource names, and deployed infrastructure as `Needs your decision` unless explicitly authorized.

