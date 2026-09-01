# Terraform Provider Contract — Zenarmor

## Purpose

`terraform-provider-zenarmor` provides declarative Terraform management of the Zenarmor NGFW and Layer 7 security capabilities running as a plugin on OPNsense.

The provider is intentionally scoped to the **Zenarmor Layer 7 policy and enforcement plane**.

General OPNsense infrastructure and Layer 3/4 controls remain the responsibility of `terraform-provider-opnsense`.

The intended ownership model is:

    Proxmox SDN
        |
        v
    OPNsense
        |
        +-- Interfaces / VLANs
        +-- Routing
        +-- NAT
        +-- L3/L4 Firewall
        +-- Suricata IDS/IPS
        +-- Zenarmor Plugin Installation / Enablement
                |
                v
          terraform-provider-zenarmor
                |
                +-- L7 Policy Assignment
                +-- Application Control
                +-- Web Control
                +-- Security / Threat Control
                +-- Explicit Allow / Block Lists
                +-- TLS Inspection
                +-- TLS Inspection Bypasses

The primary v1 use case is deterministic, Infrastructure-as-Code managed NGFW policy for trusted and untrusted home-lab workloads, including:

- GitHub Actions self-hosted runners
- Terraform/OpenTofu agents
- CI/CD workers
- Kubernetes workloads
- Servers and infrastructure services
- User VLANs
- IoT networks
- Untrusted networks
- Malware and security-testing sandboxes

A major design objective is supporting **default-deny outbound access with explicit Layer 7 allowlisting**.

---

# Provider Configuration

Example:

~~~hcl
terraform {
  required_providers {
    zenarmor = {
      source  = "khadinxc/zenarmor"
      version = "~> 1.0"
    }
  }
}

provider "zenarmor" {
  endpoint = "https://firewall.example.internal"

  api_key    = var.zenarmor_api_key
  api_secret = var.zenarmor_api_secret

  insecure_skip_verify = false
}
~~~

The provider SHOULD support:

- `endpoint`
- `api_key`
- `api_secret`
- `timeout`
- `insecure_skip_verify`

Equivalent environment variables SHOULD be supported.

Credentials and other secrets MUST:

- be marked sensitive
- never be written to logs
- never appear in diagnostics
- never be returned from data sources

Authentication MUST use a deterministic Zenarmor management interface suitable for automation.

---

# Provider Design Principles

The provider MUST:

1. Be fully declarative.
2. Support Terraform import for managed objects with stable identities.
3. Detect configuration drift.
4. Produce deterministic plans.
5. Avoid unnecessary resource replacement.
6. Treat unordered API collections as Terraform sets where appropriate.
7. Normalise API responses to prevent perpetual diffs.
8. Avoid storing ephemeral API-generated fields in configuration.
9. Resolve human-readable application and category names into internal Zenarmor identifiers.
10. Fail clearly when a configured feature is unsupported by the installed Zenarmor version or licence.
11. Support clean destruction without affecting unmanaged Zenarmor configuration.
12. Preserve configuration outside Terraform ownership.
13. Avoid modelling every individual application/category entry as an independent Terraform resource where aggregate policy resources are more appropriate.
14. Prefer stable typed Terraform attributes over arbitrary raw API payloads.
15. Never require interaction with the Zenarmor GUI for configuration represented by the v1 provider.

---

# v1 Resource Contract

The v1 provider MUST implement:

- `zenarmor_policy`
- `zenarmor_application_control`
- `zenarmor_web_control`
- `zenarmor_security_control`
- `zenarmor_exclusion`
- `zenarmor_tls_inspection`
- `zenarmor_tls_bypass`

---

# `zenarmor_policy`

Defines a Zenarmor policy and determines **which traffic receives the associated Layer 7 controls**.

This is the primary policy container to which other Zenarmor resources are attached.

Example:

~~~hcl
resource "zenarmor_policy" "github_runner" {
  name        = "github-runner"
  description = "Restricted egress policy for GitHub Actions runners."

  enabled  = true
  priority = 10

  interfaces = [
    "vtnet2",
  ]

  vlans = [
    21,
  ]

  source_networks = [
    "10.20.21.0/24",
  ]
}
~~~

## Required Capabilities

Core attributes:

- `name`
- `description`
- `enabled`
- `priority`

Traffic selectors SHOULD include where supported:

- `interfaces`
- `vlans`
- `source_networks`
- `source_addresses`
- `mac_addresses`
- `users`
- `groups`

Conditional matching SHOULD include where supported:

- `schedule`

## Behaviour

Policy priority MUST be represented deterministically.

Changing policy order SHOULD result in an in-place update where supported.

The resource MUST expose the persistent Zenarmor policy identifier through:

~~~hcl
zenarmor_policy.example.id
~~~

Dependent resources MUST be capable of referencing this identifier.

---

# `zenarmor_application_control`

Defines Layer 7 application policy.

Applications represent Zenarmor-detected protocols, services and applications rather than simple TCP or UDP ports.

Example:

~~~hcl
resource "zenarmor_application_control" "github_runner" {
  policy_id = zenarmor_policy.github_runner.id

  default_action = "block"

  allowed_applications = [
    "GitHub",
    "GitHub Actions",
    "Microsoft Azure",
  ]

  blocked_categories = [
    "Peer-to-Peer",
    "Proxy",
    "Remote Access",
    "VPN",
  ]
}
~~~

## Required Capabilities

Association:

- `policy_id`

Default behaviour:

- `default_action`

Supported default actions:

- `allow`
- `block`

Application targeting:

- `allowed_applications`
- `blocked_applications`

Application-category targeting:

- `allowed_categories`
- `blocked_categories`

## Behaviour

Applications and categories SHOULD be configurable using human-readable names.

Users MUST NOT be required to hard-code undocumented Zenarmor database identifiers.

The provider MUST resolve names to Zenarmor identifiers internally where required.

Application and category collections MUST be order-independent.

Conflicting configuration, such as the same application appearing in both allowed and blocked collections, MUST fail validation.

---

# `zenarmor_web_control`

Defines HTTP/HTTPS web filtering behaviour.

Example:

~~~hcl
resource "zenarmor_web_control" "github_runner" {
  policy_id = zenarmor_policy.github_runner.id

  profile = "custom"

  blocked_categories = [
    "Advertisements",
    "Gambling",
    "Malware",
    "Phishing",
    "Proxy",
    "Social Networks",
  ]
}
~~~

## Required Capabilities

Association:

- `policy_id`

Policy profile:

- `profile`

Where supported by Zenarmor, profiles SHOULD map directly to native profiles such as:

- `permissive`
- `moderate`
- `high`
- `custom`

Custom filtering:

- `allowed_categories`
- `blocked_categories`

## Behaviour

Web categories SHOULD be referenced using human-readable names.

Internal Zenarmor category identifiers MUST be resolved automatically.

Category collections MUST be order-independent.

Unsupported categories MUST produce actionable diagnostics.

---

# `zenarmor_security_control`

Defines security and threat-prevention controls associated with a Zenarmor policy.

Example:

~~~hcl
resource "zenarmor_security_control" "github_runner" {
  policy_id = zenarmor_policy.github_runner.id

  malware         = "block"
  phishing        = "block"
  command_control = "block"
  spyware         = "block"

  dns_over_https = "block"
  dns_over_tls   = "block"
}
~~~

## Required Capabilities

The exact schema MUST ultimately be derived from supported Zenarmor functionality.

The resource SHOULD expose first-class controls for capabilities including, where available:

- malware
- phishing
- command-and-control
- botnets
- spyware
- malicious destinations
- suspicious domains
- newly registered domains
- DNS over HTTPS
- DNS over TLS
- anonymisers
- proxy avoidance
- threat-intelligence categories

Security controls SHOULD use predictable action semantics such as:

- `allow`
- `block`
- `default`

The provider SHOULD use typed attributes rather than exposing arbitrary JSON or undocumented API payloads.

Where the installed Zenarmor version does not support a particular security capability, configuration of that capability MUST fail explicitly rather than being silently ignored.

---

# `zenarmor_exclusion`

Defines explicit Layer 7 allowlist or blocklist entries.

This resource is a primary mechanism for supporting default-deny workload egress.

Example:

~~~hcl
resource "zenarmor_exclusion" "github" {
  policy_id = zenarmor_policy.github_runner.id

  type = "allow"

  domains = [
    "github.com",
    "api.github.com",
    "objects.githubusercontent.com",
  ]

  description = "GitHub endpoints required by CI runners."
}
~~~

Global exclusions SHOULD also be supported where Zenarmor exposes global scope.

~~~hcl
resource "zenarmor_exclusion" "global_services" {
  scope = "global"

  type = "allow"

  domains = [
    "example.internal",
  ]
}
~~~

## Required Capabilities

Scope:

- `scope`

Supported scopes:

- `policy`
- `global`

Association:

- `policy_id`

Entry type:

- `type`

Supported types:

- `allow`
- `block`

Metadata:

- `description`

Destination selectors SHOULD include:

- `domains`
- `ip_addresses`
- `networks`

Where supported, additional selectors SHOULD include:

- `applications`
- `application_categories`
- `web_categories`

## Behaviour

For:

~~~text
scope = "policy"
~~~

`policy_id` MUST be required.

For:

~~~text
scope = "global"
~~~

`policy_id` MUST not be required.

Ambiguous combinations MUST be rejected during validation.

Domain, IP and network collections MUST be order-independent.

The same destination appearing simultaneously in conflicting allow and block configuration MUST fail validation where the conflict can be determined client-side.

---

# `zenarmor_tls_inspection`

Defines TLS inspection and decryption policy.

Example:

~~~hcl
resource "zenarmor_tls_inspection" "github_runner" {
  policy_id = zenarmor_policy.github_runner.id

  enabled = true
  mode    = "full"

  inspect_all = true

  block_ech = true
}
~~~

Selective inspection SHOULD be supported where Zenarmor exposes the capability.

~~~hcl
resource "zenarmor_tls_inspection" "users" {
  policy_id = zenarmor_policy.users.id

  enabled = true
  mode    = "full"

  inspect_categories = [
    "Cloud Storage",
    "Software Downloads",
  ]
}
~~~

## Required Capabilities

Association:

- `policy_id`

Core configuration:

- `enabled`
- `mode`

Inspection modes SHOULD map directly to Zenarmor functionality.

Potential modes include:

- `lightweight`
- `full`

Traffic selection SHOULD include where available:

- `inspect_all`
- `inspect_domains`
- `inspect_categories`
- `inspect_applications`

TLS privacy/evasion behaviour SHOULD include where supported:

- `block_ech`
- `exclude_certificate_pinned`
- `exclude_flows_without_dns`

## Behaviour

The provider MUST NOT pretend unsupported TLS inspection functionality exists.

If full TLS inspection requires functionality unavailable under the installed Zenarmor edition or licence, Terraform MUST fail with an actionable capability diagnostic.

---

# `zenarmor_tls_bypass`

Defines destinations, applications or categories that bypass TLS decryption.

Example:

~~~hcl
resource "zenarmor_tls_bypass" "certificate_pinning" {
  domains = [
    "pinned-service.example.com",
  ]

  description = "Application uses certificate pinning."
}
~~~

Policy-specific bypasses SHOULD be supported where Zenarmor exposes this distinction.

~~~hcl
resource "zenarmor_tls_bypass" "runner_exception" {
  policy_id = zenarmor_policy.github_runner.id

  domains = [
    "example.com",
  ]

  description = "Runner service incompatible with TLS interception."
}
~~~

## Required Capabilities

Metadata:

- `description`

Scope and association SHOULD support where available:

- `scope`
- `policy_id`

Selectors SHOULD include where supported:

- `domains`
- `applications`
- `application_categories`
- `ip_addresses`
- `networks`

All selector collections MUST be order-independent.

---

# v1 Data Source Contract

The v1 provider MUST implement sufficient discovery data sources to construct policies without requiring users to know internal Zenarmor identifiers.

Required data sources:

- `zenarmor_status`
- `zenarmor_application`
- `zenarmor_applications`
- `zenarmor_application_category`
- `zenarmor_application_categories`
- `zenarmor_web_category`
- `zenarmor_web_categories`
- `zenarmor_security_category`
- `zenarmor_security_categories`
- `zenarmor_interface`
- `zenarmor_interfaces`

---

# `zenarmor_status`

Returns information about the connected Zenarmor installation.

Example:

~~~hcl
data "zenarmor_status" "current" {}
~~~

Expected output SHOULD include where available:

- `version`
- `engine_version`
- `engine_status`
- `application_database_version`
- `threat_database_version`
- `edition`
- `license_status`
- `supported_features`
- `tls_inspection_supported`
- `full_tls_inspection_supported`
- `cloud_access_supported`

The provider SHOULD use this information internally for capability detection and validation.

Example:

~~~hcl
output "zenarmor_version" {
  value = data.zenarmor_status.current.version
}
~~~

---

# `zenarmor_application`

Looks up a single Zenarmor application.

Example:

~~~hcl
data "zenarmor_application" "github" {
  name = "GitHub"
}
~~~

Expected attributes:

- `id`
- `name`
- `category`
- `description`

Where available:

- `protocol`
- `risk`
- `tags`

Lookup by stable Zenarmor identifier SHOULD also be supported where practical.

---

# `zenarmor_applications`

Returns or searches applications known to the connected Zenarmor installation.

Example:

~~~hcl
data "zenarmor_applications" "all" {}
~~~

Filtering SHOULD be supported where practical.

Example:

~~~hcl
data "zenarmor_applications" "development" {
  category = "Software Development"
}
~~~

Returned application objects SHOULD expose:

- `id`
- `name`
- `category`
- `description`

Where available:

- `protocol`
- `risk`
- `tags`

---

# `zenarmor_application_category`

Looks up a single application category.

Example:

~~~hcl
data "zenarmor_application_category" "remote_access" {
  name = "Remote Access"
}
~~~

Expected attributes:

- `id`
- `name`
- `description`

---

# `zenarmor_application_categories`

Returns all application categories known to the installed Zenarmor application database.

Example:

~~~hcl
data "zenarmor_application_categories" "all" {}
~~~

Returned categories SHOULD expose:

- `id`
- `name`
- `description`

---

# `zenarmor_web_category`

Looks up a single web filtering category.

Example:

~~~hcl
data "zenarmor_web_category" "malware" {
  name = "Malware"
}
~~~

Expected attributes:

- `id`
- `name`
- `description`

---

# `zenarmor_web_categories`

Returns all supported web filtering categories.

Example:

~~~hcl
data "zenarmor_web_categories" "all" {}
~~~

Returned categories SHOULD expose:

- `id`
- `name`
- `description`

---

# `zenarmor_security_category`

Looks up an individual security or threat category.

Example:

~~~hcl
data "zenarmor_security_category" "command_control" {
  name = "Command and Control"
}
~~~

Expected attributes:

- `id`
- `name`
- `description`

Where available:

- `severity`
- `risk`
- `category_type`

---

# `zenarmor_security_categories`

Returns all security and threat categories exposed by the connected Zenarmor installation.

Example:

~~~hcl
data "zenarmor_security_categories" "all" {}
~~~

Returned categories SHOULD expose:

- `id`
- `name`
- `description`

Where available:

- `severity`
- `risk`
- `category_type`

---

# `zenarmor_interface`

Looks up an interface visible to Zenarmor.

Example:

~~~hcl
data "zenarmor_interface" "automation" {
  name = "vtnet2"
}
~~~

Expected attributes SHOULD include:

- `id`
- `name`
- `description`
- `enabled`
- `monitored`

Where available:

- `vlans`
- `interface_type`
- `addresses`
- `networks`

---

# `zenarmor_interfaces`

Returns interfaces available for Zenarmor inspection.

Example:

~~~hcl
data "zenarmor_interfaces" "all" {}
~~~

This data source SHOULD allow Terraform configurations to validate that policies reference interfaces actually available to or monitored by Zenarmor.

---

# Default-Deny Workload Contract

A primary v1 configuration pattern MUST support restrictive workload egress such as:

~~~hcl
resource "zenarmor_policy" "github_runner" {
  name        = "github-runner"
  description = "Restricted L7 policy for self-hosted GitHub runners."

  source_networks = [
    "10.20.21.0/24",
  ]

  priority = 10
}

resource "zenarmor_application_control" "github_runner" {
  policy_id = zenarmor_policy.github_runner.id

  default_action = "block"

  allowed_applications = [
    "GitHub",
    "Microsoft Azure",
  ]

  blocked_categories = [
    "Peer-to-Peer",
    "Proxy",
    "Remote Access",
    "VPN",
  ]
}

resource "zenarmor_web_control" "github_runner" {
  policy_id = zenarmor_policy.github_runner.id

  profile = "custom"

  blocked_categories = [
    "Malware",
    "Phishing",
    "Proxy",
  ]
}

resource "zenarmor_security_control" "github_runner" {
  policy_id = zenarmor_policy.github_runner.id

  malware         = "block"
  phishing        = "block"
  command_control = "block"

  dns_over_https = "block"
  dns_over_tls   = "block"
}

resource "zenarmor_exclusion" "github" {
  policy_id = zenarmor_policy.github_runner.id

  type = "allow"

  domains = [
    "github.com",
    "api.github.com",
    "objects.githubusercontent.com",
  ]
}

resource "zenarmor_tls_inspection" "github_runner" {
  policy_id = zenarmor_policy.github_runner.id

  enabled = true
  mode    = "full"

  inspect_all = true
}
~~~

The resulting security model is:

    Workload
       |
       v
    OPNsense
       |
       +-- L3/L4 default-deny policy
       |
       +-- Suricata IDS/IPS
       |
       +-- Zenarmor
              |
              +-- L7 default deny
              +-- Explicit required applications
              +-- Explicit destination allowlists
              +-- Web category filtering
              +-- Threat prevention
              +-- Encrypted DNS bypass prevention
              +-- TLS inspection
              +-- Explicit TLS bypasses

---

# Import Contract

Every managed resource MUST support Terraform import where Zenarmor exposes a persistent identity.

Examples:

~~~bash
terraform import zenarmor_policy.github_runner 42
~~~

Composite identifiers MAY be used where the underlying Zenarmor configuration is inherently scoped to another object.

Example:

~~~bash
terraform import zenarmor_application_control.github_runner policy/42/application-control
~~~

Import MUST populate sufficient state so that a subsequent:

~~~bash
terraform plan
~~~

accurately identifies differences between imported configuration and Terraform configuration.

Import MUST NOT require GUI interaction.

---

# Drift Detection

The provider MUST detect out-of-band modifications made through:

- the Zenarmor UI
- the Zenarmor management API
- Zenconsole where applicable
- other automation
- manual administrative changes

A refresh followed by `terraform plan` MUST surface meaningful drift.

Where Zenarmor returns collections in inconsistent order, the provider MUST normalise those collections to prevent perpetual diffs.

Computed values that change without representing configuration drift MUST not create unnecessary plans.

---

# Validation

The provider SHOULD perform client-side validation wherever the required information is available.

Validation SHOULD include:

- invalid CIDR notation
- invalid IP addresses
- invalid domains where deterministically identifiable
- invalid policy scope
- missing `policy_id`
- conflicting policy selectors
- conflicting allow/block entries
- unsupported actions
- invalid TLS inspection modes
- duplicate entries
- unknown applications
- unknown application categories
- unknown web categories
- unknown security categories
- unavailable interfaces
- interfaces not monitored by Zenarmor
- unsupported features for the connected Zenarmor version
- unsupported features for the installed licence/edition

Server-side Zenarmor validation remains authoritative.

---

# Capability Detection

Zenarmor functionality may vary according to:

- Zenarmor version
- Zenarmor engine version
- OPNsense version
- Zenarmor edition
- licence/subscription tier
- enabled modules
- platform capabilities

The provider MUST perform capability detection rather than assuming that all schema features are universally available.

Unsupported functionality MUST produce an actionable Terraform diagnostic.

Example:

    Error: TLS inspection capability unavailable

    zenarmor_tls_inspection.github_runner requests full TLS inspection,
    however the connected Zenarmor installation does not advertise this
    capability.

    Connected Zenarmor version: X.Y.Z
    Edition: <edition>

The provider MUST NOT silently ignore unsupported configuration.

The provider SHOULD expose discovered capabilities through `zenarmor_status`.

---

# State and Sensitive Data

Terraform state MUST contain only information necessary to manage declarative Zenarmor configuration.

The provider MUST NOT unnecessarily store:

- plaintext API credentials
- temporary authentication tokens
- API sessions
- decrypted TLS contents
- packet contents
- historical session data
- reporting data
- transient connection information

TLS private keys MUST NOT be introduced into state unless a future explicit certificate-management resource requires them and the security implications have been deliberately accepted.

Sensitive provider attributes MUST be marked appropriately.

---

# Error Handling

Zenarmor API errors MUST be translated into useful Terraform diagnostics.

Diagnostics SHOULD identify:

- Terraform resource
- operation being performed
- relevant Zenarmor object
- relevant identifier
- safe Zenarmor error information

Example:

    Error: Unable to update Zenarmor policy

      with zenarmor_policy.github_runner,
      on firewall.tf line 14:

    Zenarmor rejected the requested policy configuration.

Raw secrets, tokens and authentication material MUST never appear in diagnostic output.

---

# API Compatibility Contract

The provider MUST use a deterministic management interface capable of supporting Terraform CRUD semantics.

Before v1.0 release, development MUST establish that the interfaces required for every v1 resource and data source are sufficiently stable for production Terraform management.

Undocumented internal UI endpoints MAY be investigated during development but MUST NOT become part of the production v1 transport implementation unless:

1. no supported automation interface exists;
2. the endpoint behaviour is deterministic;
3. authentication can be performed securely;
4. CRUD operations can be implemented reliably;
5. stable persistent identifiers exist or can be derived;
6. behaviour can be validated across supported Zenarmor versions;
7. API response normalisation can prevent perpetual Terraform diffs;
8. compatibility can reasonably be maintained;
9. use of the endpoint does not depend on browser automation or UI scraping.

Browser automation and GUI scraping are explicitly unacceptable provider implementations.

If a required v1 capability cannot be reliably automated, that capability is a blocker for v1 rather than justification for an unreliable implementation.

---

# Terraform Lifecycle Behaviour

Resources SHOULD support normal Terraform lifecycle semantics.

Create MUST:

- create only the requested Zenarmor object
- return a stable identifier
- read the resulting configuration into Terraform state

Read MUST:

- retrieve current configuration
- normalise API values
- detect deletion outside Terraform
- never mutate remote configuration

Update MUST:

- modify resources in place where supported
- avoid replacement when Zenarmor supports mutation
- preserve fields outside Terraform ownership where appropriate

Delete MUST:

- remove only the Terraform-managed object
- avoid cascading deletion of unrelated policies or configuration unless that behaviour is an unavoidable Zenarmor API constraint
- return successful removal only after the object no longer exists

---

# OPNsense Provider Boundary

`terraform-provider-zenarmor` MUST NOT become a general-purpose OPNsense provider.

The following remain outside its ownership:

- interface creation
- VLAN creation
- bridge creation
- routing
- gateways
- DHCP
- DNS
- NAT
- L3/L4 firewall rules
- OPNsense aliases
- Suricata installation
- Suricata configuration
- Zenarmor plugin installation
- initial Zenarmor service enablement
- general OPNsense system administration

These capabilities belong to:

    terraform-provider-opnsense

The intended provisioning sequence is:

    terraform-provider-opnsense
        |
        +-- Configure interfaces/VLANs
        +-- Configure routing
        +-- Configure L3/L4 firewall
        +-- Configure Suricata
        +-- Install Zenarmor plugin
        +-- Enable/bootstrap Zenarmor
        +-- Establish API access
                     |
                     v
           terraform-provider-zenarmor
                     |
                     +-- Configure L7 policies

The OPNsense provider SHOULD eventually be capable of bootstrapping whatever authentication mechanism is required by `terraform-provider-zenarmor`.

---

# Explicitly Out of Scope for v1

The following are NOT required for v1.0:

- Zenarmor reporting dashboards
- historical reporting queries
- live session querying
- connection/session resources
- alert querying
- event querying
- packet capture
- Elasticsearch/OpenSearch reporting configuration
- Zenconsole fleet management
- multi-firewall orchestration
- CASB/cloud application action controls
- SASE configuration
- ZTNA configuration
- TLS CA lifecycle management
- certificate deployment to endpoint devices
- endpoint trust-store management
- Proxmox configuration
- Proxmox SDN configuration
- OPNsense installation
- generic OPNsense configuration
- Suricata configuration
- automatic client certificate installation
- automatic workload proxy configuration

These capabilities MAY be introduced in later provider versions where they fit the Zenarmor ownership boundary.

---

# Potential Post-v1 Resources

The provider architecture SHOULD avoid preventing future resources such as:

    zenarmor_cloud_access_control
    zenarmor_tls_ca
    zenarmor_schedule
    zenarmor_reporting_configuration

Potential future data sources MAY include:

    zenarmor_sessions
    zenarmor_events
    zenarmor_threats
    zenarmor_statistics

These are not required for v1.

---

# Final v1 Resource Surface

Resources:

    zenarmor_policy
    zenarmor_application_control
    zenarmor_web_control
    zenarmor_security_control
    zenarmor_exclusion
    zenarmor_tls_inspection
    zenarmor_tls_bypass

Data sources:

    zenarmor_status

    zenarmor_application
    zenarmor_applications

    zenarmor_application_category
    zenarmor_application_categories

    zenarmor_web_category
    zenarmor_web_categories

    zenarmor_security_category
    zenarmor_security_categories

    zenarmor_interface
    zenarmor_interfaces

---

# v1 Acceptance Criteria

`terraform-provider-zenarmor` v1.0 is functionally complete when a user can take an already installed and API-accessible Zenarmor instance and configure the required NGFW Layer 7 policy entirely through Terraform.

For a workload such as a self-hosted GitHub Actions runner, Terraform MUST be capable of declaratively defining:

1. which interface, VLAN, network or workload receives a policy;
2. deterministic policy priority;
3. a default-deny application posture;
4. explicitly permitted applications;
5. explicitly denied applications;
6. permitted and denied application categories;
7. permitted and denied web categories;
8. malware and threat-prevention controls;
9. command-and-control protection;
10. encrypted DNS bypass prevention;
11. explicit destination domain allowlists;
12. explicit destination domain blocklists;
13. IP/network exclusions where supported;
14. TLS inspection;
15. TLS inspection targeting;
16. TLS inspection exceptions;
17. discovery of valid applications;
18. discovery of valid application categories;
19. discovery of valid web categories;
20. discovery of valid security categories;
21. discovery of Zenarmor-monitored interfaces;
22. capability detection against the connected installation;
23. import of existing Zenarmor configuration;
24. detection of configuration drift;
25. deterministic Terraform plans;
26. safe update and destruction;
27. actionable diagnostics for unsupported functionality.

A normal operational workflow MUST be:

~~~bash
terraform init
terraform plan
terraform apply
~~~

No manual interaction with the Zenarmor GUI SHOULD be required after the initial Zenarmor installation, service bootstrap and API credential provisioning.

---

# Architectural End State

The complete home-lab firewall ownership model is:

    Terraform
       |
       +-- terraform-provider-proxmox
       |      |
       |      +-- Compute
       |      +-- Virtual Networks
       |      +-- VLAN-aware infrastructure
       |
       +-- terraform-provider-opnsense
       |      |
       |      +-- Interfaces
       |      +-- VLANs
       |      +-- Routing
       |      +-- NAT
       |      +-- L3/L4 Firewall
       |      +-- Suricata IDS/IPS
       |      +-- Zenarmor installation
       |      +-- Zenarmor bootstrap
       |
       +-- terraform-provider-zenarmor
              |
              +-- L7 policy assignment
              +-- Application identification/control
              +-- Web filtering
              +-- Threat prevention
              +-- Explicit L7 allowlists
              +-- Explicit L7 blocklists
              +-- TLS inspection
              +-- TLS bypasses

The resulting enforcement architecture is:

    Workload
       |
       v
    Proxmox SDN
       |
       v
    OPNsense
       |
       +-- Routing / segmentation
       |
       +-- L3/L4 firewall
       |
       +-- Suricata IDS/IPS
       |
       v
    Zenarmor
       |
       +-- Identify application
       +-- Evaluate L7 policy
       +-- Evaluate web category
       +-- Evaluate threat intelligence
       +-- Enforce allow/block policy
       +-- Inspect TLS where configured
       |
       v
    Internet / external service

The strategic target is that sensitive infrastructure workloads can operate under:

    deny by default
          +
    explicitly required network flows
          +
    explicitly required applications
          +
    explicitly required destinations
          +
    threat prevention
          +
    TLS visibility where appropriate

without requiring ongoing manual firewall administration.

This defines the functional end-state contract for `terraform-provider-zenarmor` v1.0.