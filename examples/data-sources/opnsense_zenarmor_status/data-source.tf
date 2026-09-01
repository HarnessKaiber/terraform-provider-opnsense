data "opnsense_zenarmor_status" "current" {}

output "zenarmor_plugin_version" {
  value = data.opnsense_zenarmor_status.current.plugin_version
}
