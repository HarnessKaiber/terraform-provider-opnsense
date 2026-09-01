# Zenarmor Local API Compatibility

## Selected interface

The provider targets Zenarmor management interfaces installed locally on
OPNsense and authenticates through the OPNsense API using an API key and secret.
It does not depend on Zenconsole or Zenarmor's separately licensed native API.

OPNsense API routes have the form
`/api/<module>/<controller>/<command>` and use the API key and secret as HTTP
Basic Auth username and password.

## Discovery state

| Surface | State |
|---|---|
| OPNsense API authentication | Verified against the live lab firewall |
| Installed Zenarmor package and version | `os-sensei` 2.6.2 verified through firmware inventory |
| Zenarmor OPNsense MVC controllers | `GET /api/zenarmor/policy` and `GET /api/zenarmor/status` verified |
| Stable local policy create | Verified through `POST /api/zenarmor/policy` with a unique name |
| Stable local policy update/delete | Blocked by the Zenarmor 2.6.2 API ACL |
| Runtime status | Agent, engine, database, licence, deployment and cloud-intelligence fields verified |
| Monitored interface catalogue and exact lookup | Verified through `interfaces_list` in local status response |
| Application/web/security catalogues and TLS controls | Candidate local routes return 404; not yet proven |

The public OPNsense `vendor/sunnyvalley` plugin contains the Sunny Valley package
repository bootstrap rather than the proprietary Zenarmor controller sources.
Consequently, endpoint discovery is verified directly against the installed
package on the test firewall.

### Verified policy-list endpoint

`GET /api/zenarmor/policy` accepts the normal OPNsense API key and secret and
returns a `policies` collection. Observed stable fields are `id`, `local_id`,
`cloud_policyid`, `name`, `isCentralized`, `isActive`, `isDefault`, `user`,
`nodes`, `tags`, `projects`, and `checksum`.

The endpoint is implemented in `opnsense-go/pkg/zenarmor` and validated by the
`opnsense_zenarmor_policies` Terraform data source using a separate live client
call.

`POST /api/zenarmor/policy` was tested with uniquely named disabled disposable
policies. The controller creates a policy from the submitted `name`, assigns a
persistent integer ID, seeds default allow controls, applies the configuration,
and returns the new ID.

The published `os-sensei-2.6.2.pkg` contains
`OPNsense/Zenarmor/Api/PolicyController.php`. Its `deleteAction()` performs the
required cleanup and expects a POST field named `cloud_id`. However, the
`ACL.xml` shipped in the same package permits `api/zenarmor/policy` and
`api/zenarmor/policy/application_database` only. It does not permit
`api/zenarmor/policy/delete` or the other policy action routes. Live requests to
the delete action consequently return HTTP 403 before controller execution.

The collection controller's PATCH branch is a no-op success response. Posting a
full document with an existing `id` to the collection route creates another
policy instead of updating the existing object. Terraform policy CRUD therefore
cannot be implemented safely against stock Zenarmor 2.6.2: destroy would either
fail or orphan configuration. The vendor must expose the delete, detail,
configuration, name, and apply actions through a supported API ACL.

### Verified status endpoint

`GET /api/zenarmor/status` accepts the same OPNsense credentials. The observed
2.6.2 response supplies agent installation/state/version, engine version and
state, application database version, database service state/version, licence,
deployment mode, cloud-threat-intelligence state, and monitored interface data.
The status data source maps only stable typed fields. Missing version-dependent
fields become null, collection ordering is normalised, and TLS capability stays
false unless a future supported controller positively advertises it.
The monitored interface collection backs deterministic list and exact-name
Terraform data sources; interface and tag ordering are normalised before state
is written.

Read-only discovery also tested likely catalogue and interface controller names.
Those candidates returned 404, while `/api/zenarmor/configuration` returned 403.
Neither result is treated as an implemented management surface.

Discovery starts with known read-only OPNsense endpoints and installed-package
inspection. Unknown controller commands are not blindly invoked: OPNsense routes
actions by controller method, and an incorrectly guessed GET can be unsafe on
older or vulnerable implementations.

## Acceptance rule

A local interface may back a provider surface only after testing proves stable
authentication, identifiers, read/create/update/delete semantics, response
normalisation, validation and configuration reload behavior. Browser session
replay, GUI scraping, direct proprietary database edits and licence-control
circumvention are prohibited.

If Zenarmor exposes no deterministic local management interface for a contracted
surface, that surface remains blocked rather than falling back to the paid native
API or pretending support exists.
