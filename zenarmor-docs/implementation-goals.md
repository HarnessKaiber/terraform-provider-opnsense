# Zenarmor Provider Implementation Goals

This document turns the [provider contract](provider-contract.md) into an
ordered implementation plan. It is the source of truth for what is complete,
what is next, and which dependencies must be satisfied before work starts.

The provider initially targets the Zenarmor plugin running on OPNsense. Work is
deliberately completed one provider surface at a time. A goal is not complete
when its schema compiles; it is complete only after its contract behaviour has
been exercised end to end against a real Zenarmor installation.

## Status legend

- `[ ]` Not started
- `[~]` In progress
- `[x]` Complete
- `[!]` Blocked, with the blocker recorded under the goal

## Rules for every implementation goal

Every resource and data source must meet all applicable gates below before its
goal can be marked complete.

1. The supported Zenarmor API endpoints and version/licence constraints are
   recorded from observed behaviour, not inferred from the GUI.
2. The Go client has deterministic request/response types, context-aware calls,
   safe diagnostics, and unit tests for normalisation and error handling.
3. The Terraform schema uses typed attributes, sets for unordered collections,
   client-side validation where deterministic, and no secret or transient state.
4. Provider registration, documentation, and a usable example are present.
5. A Terraform acceptance test runs against a newly built OPNsense image with
   Zenarmor installed and bootstrapped. It is based on real provider use, not a
   connectivity or schema smoke test.
6. The acceptance test validates Terraform's result using a separate Zenarmor
   API call. Provider state alone is never the independent oracle.
7. A resource test covers create, read, in-place update where supported, import,
   API-verified out-of-band drift, removal outside Terraform, and destroy. It
   confirms that unmanaged configuration is preserved and that a final plan is
   empty.
8. A data-source test queries Terraform and the Zenarmor API separately, compares
   stable fields and collection membership, tests filtering/lookup failures where
   applicable, and confirms a repeated plan is empty.
9. Unsupported version, edition, licence, or feature combinations return an
   actionable diagnostic and are not silently ignored.
10. Unit tests, acceptance tests, formatting, linting, documentation generation,
    and the full existing suite pass.

Tests must create uniquely named fixtures and clean them up. They must not depend
on mutable public data, execution order, pre-existing user configuration, or the
contents of another test. Tests that cannot run on the pipeline image must fail
as an explicitly documented capability blocker; silently skipping a contracted
surface does not satisfy the goal.

## Dependency order

```text
G00 contract/API discovery
 |
 +--> G01 provider and client foundation
 |     |
 |     +--> G02 hermetic OPNsense + Zenarmor test platform
 |     |     |
 |     |     +--> G03 status/capabilities
 |     |           |
 |     |           +--> G04-G11b discovery data sources
 |     |           |     |
 |     |           |     +--> G13 application control
 |     |           |     +--> G14 web control
 |     |           |     +--> G15 security control
 |     |           |     +--> G16 exclusions
 |     |           |     +--> G17 TLS inspection
 |     |           |     +--> G18 TLS bypass
 |     |           |
 |     |           +--> G12 policy
 |     |                 |
 |     |                 +--> G13-G18 policy-scoped controls
 |     |
 |     +--> local testing through creds.yaml
 |
 +--> G19 default-deny composition and v1 qualification
```

Later goals may refine an earlier schema only when observed API behaviour makes
that necessary. They must not bypass an incomplete dependency.

## Foundation goals

### [~] G00 — Validate the automation contract against Zenarmor

**Depends on:** none

**Goal:** Establish that every v1 surface has a stable, non-browser automation
interface capable of Terraform CRUD or read semantics.

**Deliverables and completion gates:**

**Current state:** OPNsense authentication, firmware inventory, and the local
policy-list controller are verified. Remaining resource and catalogue endpoints
still require deterministic lifecycle discovery. See
[API compatibility](api-compatibility.md).

- Inventory installed OPNsense, Zenarmor plugin, engine, database, edition, and
  licence versions available in the local test target.
- Probe authentication and candidate endpoints using `creds.yaml`; redact all
  credentials, tokens, cookies, and sensitive response fields from output and
  fixtures.
- Record endpoints, methods, identities, concurrency/reconfigure requirements,
  error shapes, defaults, collection ordering, and capability restrictions for
  all seven resources and eleven data sources.
- Prove create/read/update/delete semantics and preservation of unrelated fields
  for each managed object before implementing its Terraform resource.
- Identify any contract item without a reliable API as a v1 blocker. GUI or
  browser automation is not an acceptable fallback.

### [~] G01 — Provider and Zenarmor API client foundation

**Depends on:** G00 authentication and transport findings

**Goal:** Create the readable Go/provider structure used by every later surface,
following the conventions in `terraform-provider-opnsense` without copying its
OPNsense ownership into this provider.

**Deliverables and completion gates:**

- Go module, provider entry point, provider factory, service packages, shared
  validators, test harness, examples, generated-doc structure, and build/lint
  configuration.
- Provider attributes `endpoint`, `api_key`, `api_secret`, `timeout`, and
  `insecure_skip_verify`, plus documented `ZENARMOR_*` environment equivalents.
  The key and secret authenticate to the local OPNsense API.
- Sensitive credentials never enter Terraform state, diagnostics, or logs.
- HTTP client with bounded timeout, TLS configuration, authentication, safe
  errors, response-size handling, and deterministic retry behaviour only where
  safe.
- Unit tests for provider configuration, environment fallback, redaction,
  transport errors, server errors, and cancellation.
- A local acceptance runner reads `creds.yaml` at execution time, maps it into
  environment variables, and never generates or commits Terraform containing
  credentials. The file is ignored by Git.

### [!] G02 — Hermetic OPNsense and Zenarmor acceptance-test platform

**Depends on:** G00 plugin/bootstrap findings; G01

**Goal:** Build the OPNsense image in every PR pipeline, install and bootstrap
Zenarmor, create automation credentials, and run deterministic Terraform tests.

**Deliverables and completion gates:**

**Current blocker:** The image recipe must install and bootstrap the local
Zenarmor package, after which the same OPNsense API credential provisioning used
by `terraform-provider-opnsense` will be reused. No credentials will be baked
into an image artifact.

- Reusable image workflow adapted from `terraform-provider-opnsense`; the image
  version, checksum, Zenarmor plugin version, and installation inputs are pinned.
- The image is actually produced by the workflow. Caching may accelerate an
  identical pinned build but must not replace the declared build recipe or allow
  an unverified mutable image.
- Zenarmor installation, first-run bootstrap, engine readiness, API credential
  provisioning, and test fixtures are fully automated without GUI interaction.
- Readiness uses bounded health polling rather than a fixed boot delay alone.
- PR tests start an isolated VM, run tests serially within an instance or use one
  isolated VM per parallel shard, collect useful redacted diagnostics, and always
  tear down the VM.
- A coverage-manifest check fails CI when a registered resource or data source is
  missing its Terraform acceptance test and independent API validation.
- The pipeline verifies that no acceptance test was skipped and no test fixture
  remains after the suite.

## Discovery and capability goals

### [!] G03 — `zenarmor_status` data source and capability service

**Depends on:** G01, G02

**Goal:** Provide installation status and the shared capability model required by
all feature-dependent resources.

**Current blocker:** A deterministic local status/version/licence/capability
interface has not yet been proven on the installed plugin. The data source
remains unregistered instead of returning inferred or fabricated values.

**Contract coverage:** version, engine version/status, database versions, edition,
licence status, supported features, TLS inspection support, full TLS inspection
support, and cloud-access support where the API exposes them.

**Additional completion gates:** stable feature names and deterministic ordering;
independent API comparison; explicit handling of absent fields on older versions;
reusable capability checks that include connected version and edition in safe
diagnostics.

### [ ] G04 — `zenarmor_application_categories` data source

**Depends on:** G03

**Goal:** Return the complete, normalised application-category catalogue using
stable IDs, names, and descriptions.

**Additional completion gates:** independent API membership/count validation,
order-independent Terraform state, duplicate-ID/name detection, and empty-result
semantics.

### [ ] G05 — `zenarmor_application_category` data source

**Depends on:** G04

**Goal:** Resolve exactly one application category by human-readable name and,
where stable and practical, by ID.

**Additional completion gates:** exact unambiguous lookup, independent API field
comparison, and actionable not-found/duplicate-match diagnostics.

### [ ] G06 — `zenarmor_applications` data source

**Depends on:** G04

**Goal:** Return the normalised application catalogue and support category
filtering where the API can do so deterministically.

**Additional completion gates:** stable ID/name/category/description plus exposed
protocol/risk/tags where available; filtered and unfiltered API comparisons; set
semantics for tags and returned collections.

### [ ] G07 — `zenarmor_application` data source

**Depends on:** G06

**Goal:** Resolve exactly one application by human-readable name and, where
stable and practical, by ID.

**Additional completion gates:** independent API comparison and actionable
unknown/ambiguous-name diagnostics. The lookup logic becomes the shared resolver
used by application-aware resources.

### [ ] G08 — `zenarmor_web_categories` data source

**Depends on:** G03

**Goal:** Return every supported web-filtering category with stable ID, name, and
description.

**Additional completion gates:** independent API membership/count validation,
order-independent state, duplicate detection, and version/database normalisation.

### [ ] G09 — `zenarmor_web_category` data source

**Depends on:** G08

**Goal:** Resolve exactly one web category by human-readable name and, where
stable and practical, by ID.

**Additional completion gates:** independent API comparison and actionable
unknown/ambiguous-name diagnostics. Its resolver is shared by web and TLS
resources.

### [ ] G10a — `zenarmor_security_categories` data source

**Depends on:** G03

**Goal:** Expose the security-category catalogue with stable ID/name/description
and severity, risk, and category type where available.

**Additional completion gates:** independent API membership/count validation;
normalised ordering; duplicate detection; deterministic absent-field semantics.

### [ ] G10b — `zenarmor_security_category` data source

**Depends on:** G10a

**Goal:** Resolve exactly one security category by human-readable name and,
where stable and practical, by ID.

**Additional completion gates:** independent API field comparison; actionable
not-found/ambiguous-match diagnostics; shared resolver for security controls.

### [~] G11a — `zenarmor_interfaces` data source

**Depends on:** G03

**Goal:** Return the normalised catalogue of interfaces visible to Zenarmor.

**Contract coverage:** stable ID/name/description, enabled/monitored state, and
VLANs/type/addresses/networks where available.

**Additional completion gates:** independent API membership/count validation; the
test validates a known pipeline interface; unavailable and unmonitored interfaces
remain distinguishable; order-independent nested collections.

### [~] G11b — `zenarmor_interface` data source

**Depends on:** G11a

**Goal:** Resolve exactly one interface visible to Zenarmor by human-readable
name and, where stable and practical, by ID.

**Contract coverage:** stable ID/name/description, enabled/monitored state, and
VLANs/type/addresses/networks where available.

**Additional completion gates:** independent API field comparison; actionable
not-found/ambiguous-match diagnostics; shared validation supports policies.

## Managed-resource goals

Only one managed-resource goal is implemented at a time in the order below. A
resource is complete before implementation starts on the next resource.

### [ ] G12 — `zenarmor_policy` resource

**Depends on:** G03, G11a, G11b

**Verified blocker (Zenarmor 2.6.2):** collection POST creates a persistent
policy, but the packaged API ACL omits the existing `deleteAction`,
`detailAction`, and update/configuration actions. API-key requests to those
actions return 403. G12 resumes when a supported plugin release exposes the
complete lifecycle or the vendor supplies a supported ACL/API extension; the
provider must not remove Terraform state while leaving the remote policy behind.

**Goal:** Manage the primary policy container and establish stable policy
identities for every policy-scoped control.

**Contract coverage:** name, description, enabled, deterministic priority,
interfaces, VLANs, source networks/addresses, MAC addresses, users, groups, and
schedule wherever supported by the detected installation.

**Additional completion gates:** stable ID; in-place priority/order update;
selector validation; interface validation; collection normalisation; import;
API-induced description/priority/selector drift; external deletion; destroy that
does not affect unmanaged policies; no-op plan after each settled operation.

### [ ] G13 — `zenarmor_application_control` resource

**Depends on:** G12, G04-G07

**Goal:** Manage application policy for one policy, including the v1 default-deny
application posture.

**Contract coverage:** policy ID; allow/block default action; allowed/blocked
applications; allowed/blocked application categories.

**Additional completion gates:** names resolve internally to stable IDs; all
collections are sets; overlap is rejected client-side; unknown names are
actionable; import uses a stable or documented composite ID; independent API
checks prove default action and resolved memberships; tests cover allow-to-block
update and out-of-band membership drift without changing unrelated policy fields.

### [ ] G14 — `zenarmor_web_control` resource

**Depends on:** G12, G08, G09

**Goal:** Manage native web profiles and custom allow/block web categories for a
policy.

**Contract coverage:** policy ID, supported native profile, allowed categories,
and blocked categories.

**Additional completion gates:** profile/category compatibility validation;
name-to-ID resolution; overlap and unknown-category diagnostics; import;
independent API checks for profile and memberships; custom-profile update and
out-of-band drift; preservation of unrelated policy controls.

### [ ] G15 — `zenarmor_security_control` resource

**Depends on:** G12, G03, G10a, G10b

**Goal:** Expose only observed, stable, first-class security controls with
predictable `allow`, `block`, and `default` semantics.

**Contract coverage:** malware, phishing, command-and-control, botnets, spyware,
malicious/suspicious/newly registered destinations, encrypted DNS, anonymisers,
proxy avoidance, and threat-intelligence categories where actually supported.

**Additional completion gates:** no arbitrary JSON; per-control capability and
licence diagnostics; import; independent API validation for each supported
control family; allow/block/default updates; out-of-band drift; preservation of
unsupported or unmanaged server fields.

### [ ] G16 — `zenarmor_exclusion` resource

**Depends on:** G12, G04-G10b as required by observed selectors

**Goal:** Manage explicit policy or global allow/block entries used by
default-deny workload egress.

**Contract coverage:** scope, conditional policy ID, type, description, domains,
IP addresses, networks, and application/category selectors where supported.

**Additional completion gates:** scope/policy validation; IP/CIDR/domain
validation; order-independent selectors; client-detectable conflicts rejected;
stable or composite import ID; independent API checks for global and policy
fixtures; allow/block and selector updates; drift and external-deletion tests;
destroy removes only the managed exclusion.

### [ ] G17 — `zenarmor_tls_inspection` resource

**Depends on:** G12, G03, G04-G09 as required by observed selectors

**Goal:** Manage policy TLS inspection only to the extent advertised and proven
by the connected Zenarmor installation.

**Contract coverage:** policy ID, enabled, supported mode, inspect-all and
selective domains/categories/applications, ECH behaviour, certificate-pinning
behaviour, and flows-without-DNS behaviour where supported.

**Additional completion gates:** edition/licence/capability preflight with
actionable diagnostics; mutually exclusive selector validation; human-name
resolution; import; independent API verification of actual inspection mode and
targeting; enable/mode/target update; out-of-band drift; safe disable/destroy;
explicit negative test when full inspection is unavailable.

### [ ] G18 — `zenarmor_tls_bypass` resource

**Depends on:** G17, G04-G09 as required by observed selectors

**Goal:** Manage global and policy-specific TLS decryption exceptions without
weakening unrelated inspection configuration.

**Contract coverage:** description, supported scope and conditional policy ID,
domains, applications/categories, IP addresses, and networks.

**Additional completion gates:** selector and scope validation; set semantics;
human-name resolution; stable or composite import ID; independent API checks for
global and policy fixtures; selector/scope-supported updates; out-of-band drift;
destroy proves the parent inspection configuration and unmanaged bypasses remain.

## Release qualification

### [ ] G19 — Default-deny composition and v1 contract qualification

**Depends on:** G00-G18, including all lettered subgoals

**Goal:** Prove that the resources work together as the contract's intended
declarative Layer 7 policy, not merely as isolated CRUD implementations.

**Deliverables and completion gates:**

- Apply a complete self-hosted-runner-style configuration containing a policy,
  default-block application control, explicit applications, web and security
  controls, destination exclusions, TLS inspection, and TLS bypasses.
- Validate the composed configuration through separate API calls and, where a
  deterministic test workload is possible, observable allowed and denied traffic.
- Re-apply and require an empty plan; reorder every set-valued input and require
  an empty plan.
- Mutate representative fields out of band and require refresh/plan to report
  meaningful drift, then apply to reconcile it.
- Import every managed surface into fresh state and require the configured plan
  to be empty after normalisation.
- Destroy the composition and prove through the API that managed objects are gone
  while seeded unmanaged configuration remains.
- Run every resource and data-source test in the PR-built image pipeline with no
  skips, races, leaked fixtures, credential output, or perpetual diffs.
- Check every item in the contract's v1 acceptance criteria and record evidence
  or an explicit release blocker.

## Current execution point

`G00`, `G01`, `G03`, `G11a`, and `G11b` are in progress.
`opnsense_zenarmor_status`, `opnsense_zenarmor_policies`,
`opnsense_zenarmor_interfaces`, and `opnsense_zenarmor_interface` are registered
and have live Terraform acceptance coverage with independent OPNsense API
validation. The interface data sources use the stable `interfaces_list` field
from `/api/zenarmor/status`. Managed-resource lifecycle work remains blocked
until disposable policy create/delete semantics are proven.
