---
title: Interchangeable Branding Contract
description: Configuration and validation rules for reusable white-label solution branding
---

## Define one canonical profile

Create one canonical source or a documented source-of-truth hierarchy. Reuse the repository's established settings, Helm values, typed configuration, theme provider, or generation mechanism.

Use `BRAND_*` for presentation and product identity. Use `CUSTOMER_*` only when the application models a customer or tenant as business data.

Recommended concepts include:

```dotenv
BRAND_BUSINESS_NAME=Contoso
BRAND_LEGAL_NAME=Contoso Ltd
BRAND_PRODUCT_NAME=<existing-or-user-approved-product-name>
BRAND_SHORT_NAME=<existing-or-user-approved-short-name>
BRAND_SLUG=contoso
BRAND_DNS_SLUG=contoso
BRAND_RESOURCE_PREFIX=contoso
BRAND_DOMAIN=contoso.example
BRAND_SUPPORT_EMAIL=support@contoso.example
BRAND_PRIMARY_COLOR=#2563EB
BRAND_SECONDARY_COLOR=#0F172A
BRAND_FONT_FAMILY=Inter
BRAND_LOGO_PATH=assets/brand/logo.svg
BRAND_FAVICON_PATH=assets/brand/favicon.svg
BRAND_DEFAULT_LOCALE=en-US
```

These values illustrate configuration shape, not a proposed brand profile. Preserve existing product names and abbreviations unless product branding is explicitly in scope and the replacement is user-approved. Never derive initials, acronyms, monograms, logos, or icon treatments from an example or an inferred name. For organization-only debranding, replace or parameterize the organization identity while leaving the product identity unchanged.

Use reserved example domains. Never invent product names, production identifiers, endpoints, credentials, tenant IDs, or subscription IDs.

## Separate naming concerns

Do not assume one string can safely represent all of these concerns:

* Business display name and legal name
* Product name, short name, and package identifier
* Lowercase slug and DNS-safe slug
* Cloud-resource prefix and globally unique resource name
* Tenant identity, deployment identity, and subscription identity
* Domain, support contact, logo, icon, and theme tokens
* Package metadata (`package.json` name, Python project name, NuGet ID)
* Storage namespaces, queue prefixes, cache key prefixes, and logger names

These carry brand silently. A package named `acme-frontend` embeds a brand short name; a Redis namespace `acme/chat/` does the same. Both trip the two-profile harness when a profile value appears in them. Make them brand-neutral or derive them from canonical configuration at build time.

## Classify delivery mechanisms

Assign each value to one mechanism:

* Runtime configuration for values operators can change without rebuilding
* Build-time configuration for compiled metadata and bundled frontend values
* Deployment configuration for infrastructure and environment-specific names
* Generated assets for manifests, package metadata, documentation, logos, and icons

Document precedence when a value appears in more than one representation. Generate downstream files from the canonical profile when practical.

## Enforce naming constraints

Record allowed characters, case, maximum length, reserved names, collision behavior, provider-specific requirements, immutability, and global uniqueness. Do not silently derive infrastructure identifiers from display text.

Avoid hardcoded brand identifiers in image repositories, Kubernetes namespaces, Helm releases, ingress hosts, DNS names, resource groups, storage accounts, managed identities, package names, and public namespaces. Make each separately configurable where the platform permits it.

## Parameterize visual identity

A palette, a typeface, and a logo identify an organisation as reliably as its name, and renaming never touches them. Route them through the same canonical profile as the text values.

* Replace hardcoded palette values in stylesheets, theme providers, Tailwind or design-token configuration, inline SVG fills, chart series colours, email templates, and PDF or report styling with named tokens.
* Keep semantic tokens such as `--color-primary` and `--color-danger` separate from raw palette values, so a new profile restyles the solution without touching component code.
* Replace logo, wordmark, favicon, app icon, splash, and social-preview assets with profile-referenced paths and ship neutral placeholders.
* When replacing a visual asset with a different format, verify that the consuming code and framework accept the new format. A PNG replaced with an SVG breaks any rendering pipeline that does not allow SVG processing, such as framework image components that require explicit SVG opt-in, image optimizers that reject vector formats, or build tools that expect raster input. Test the replaced asset by rendering it through the same code path that production uses, not just by confirming the file exists on disk. Development-mode auth bypasses can hide login-page and unauthenticated-surface breakage.
* Check the remaining palette against accessibility contrast requirements after substitution, because a swapped colour can silently break a ratio the original satisfied.

Treat bundled fonts and logo files as a licensing question, not only an attribution one. Corporate typefaces are usually licensed to one organisation, and shipping the file in a public repository redistributes it. Replace them with an openly licensed family, and record the substitution.

## Validate two profiles

Validate at least two distinct fictional profiles, such as Contoso and Fabrikam:

1. Render, build, or run the relevant output with each profile.
2. Compare visible text, generated assets, manifests, package metadata, and deployment templates.
3. Confirm the palette, typeface, logo, and favicon change with the profile, then search the rendered output for the first profile's hex values and asset names. A build that swaps every string but renders in the original brand colours has not been debranded.
4. Confirm profile switching requires no source edits except documented non-configurable artifacts.
5. Search for literals from both profiles to ensure example values did not become architectural defaults.
6. Verify intentionally stable identifiers remain stable.
7. Report values that require rebuild, regeneration, redeployment, or resource replacement.

Validate at the layer that consumes the value, not the layer that carries it. A deployment manifest can pass a brand variable to a container that no code reads, or that reads it too late to matter, and rendering that manifest for both profiles proves only that templating works. Framework build-time inlining is the usual reason: a bundler that fixes a variable at build time makes the runtime copy inert, so the profile is decided when the image is built and no deployment-time override reaches the user.

Two symptoms mark a carrier that nothing consumes. Searching the codebase for the variable name finds no reader, and the carried value has drifted out of agreement with the value the application actually uses, because nothing depended on them matching. When you find this, delete the dead carrier rather than leaving it as a knob that lies, and state plainly which stage now decides the profile.

## Consume a design system

When `DESIGN_SYSTEM` names a structured system (Fluent 2, Carbon, Material, Radix, Shadcn, or similar), plan from the consuming integration boundary:

1. Identify the concrete consumer before proposing implementation. A consumer is an installed component library, theme provider, CSS variable contract, or generation API that will read the mapped tokens. If no current or approved future consumer exists, ask one focused scope question or record adoption as a separately gated unit.
2. Use the system's supported theming contract. Prefer theme factories, ramp generators, and semantic alias APIs over manual primitive assignment. Never assign a raw color to a numbered ramp slot without an official algorithm or a documented human design decision.
3. Search component code for hardcoded palette values (hex literals, RGB, HSL) that bypass the token layer. These are consumers that the new tokens must reach. Include a unit to wire them to the generated ramp.
4. Record unmapped values as gaps requiring human decisions. Map only values the repository uses and inputs the consumer requires.

Do not fetch design system documentation during apply or verify. The plan phase captures the token mapping; later phases consume it from the plan.

When `DESIGN_SYSTEM` is a visual reference, extract a proposed palette and type selection, present them to the user for approval, and record the approved values as canonical profile entries. Do not treat a visual reference as a structured token source.

## Portability acceptance criteria

Branding is portable only when:

* Business and product names come from documented canonical configuration
* User-facing text does not duplicate customer literals
* Infrastructure identifiers remain separately configurable
* Static and generated artifacts have a regeneration path
* Example values use safe fictional data and reserved domains
* Two profiles pass the relevant validation
* Remaining source edits and immutable resources are explicitly documented

