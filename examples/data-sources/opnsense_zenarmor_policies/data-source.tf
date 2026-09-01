data "opnsense_zenarmor_policies" "all" {}

output "zenarmor_policies" {
  value = data.opnsense_zenarmor_policies.all.policies
}
