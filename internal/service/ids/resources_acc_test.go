package ids_test

import (
	"fmt"
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIDSSettingsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: testAccIDSSettingsConfig("192.0.2.0/24"), Check: resource.TestCheckResourceAttr("opnsense_ids_settings.test", "enabled", "false")},
			{Config: testAccIDSSettingsConfig("198.51.100.0/24"), Check: resource.TestCheckResourceAttr("opnsense_ids_settings.test", "home_networks.#", "1")},
		},
	})
}

func TestAccIDSRulesetResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: testAccIDSRulesetConfig(false), Check: resource.TestCheckResourceAttr("opnsense_ids_ruleset.test", "enabled", "false")},
			{ResourceName: "opnsense_ids_ruleset.test", ImportState: true, ImportStateVerify: true},
			{Config: testAccIDSRulesetConfig(true), Check: resource.TestCheckResourceAttr("opnsense_ids_ruleset.test", "enabled", "true")},
		},
	})
}

func TestAccIDSPolicyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: testAccIDSPolicyConfig(1000, "acceptance policy"), Check: resource.TestCheckResourceAttr("opnsense_ids_policy.test", "priority", "1000")},
			{ResourceName: "opnsense_ids_policy.test", ImportState: true, ImportStateVerify: true},
			{Config: testAccIDSPolicyConfig(1001, "acceptance policy updated"), Check: resource.TestCheckResourceAttr("opnsense_ids_policy.test", "description", "acceptance policy updated")},
		},
	})
}

func TestAccIDSPolicyRuleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: testAccIDSPolicyRuleConfig("acceptance rule"), Check: resource.TestCheckResourceAttr("opnsense_ids_policy_rule.test", "enabled", "false")},
			{ResourceName: "opnsense_ids_policy_rule.test", ImportState: true, ImportStateVerify: true},
			{Config: testAccIDSPolicyRuleConfig("acceptance rule updated"), Check: resource.TestCheckResourceAttr("opnsense_ids_policy_rule.test", "message", "acceptance rule updated")},
		},
	})
}

func TestAccIDSUserRuleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: testAccIDSUserRuleConfig("acceptance user rule"), Check: resource.TestCheckResourceAttr("opnsense_ids_user_rule.test", "enabled", "false")},
			{ResourceName: "opnsense_ids_user_rule.test", ImportState: true, ImportStateVerify: true},
			{Config: testAccIDSUserRuleConfig("acceptance user rule updated"), Check: resource.TestCheckResourceAttr("opnsense_ids_user_rule.test", "description", "acceptance user rule updated")},
		},
	})
}

func testAccIDSSettingsConfig(homeNetwork string) string {
	return fmt.Sprintf(`
resource "opnsense_ids_settings" "test" {
  enabled       = false
  mode          = "pcap"
  interfaces    = ["wan"]
  home_networks = [%q]
  promiscuous   = false
}
`, homeNetwork)
}

func testAccIDSRulesetConfig(enabled bool) string {
	return fmt.Sprintf(`
resource "opnsense_ids_ruleset" "test" {
  filename = "abuse.ch.feodotracker.rules"
  enabled  = %t
}
`, enabled)
}

func testAccIDSPolicyConfig(priority int, description string) string {
	return fmt.Sprintf(`
resource "opnsense_ids_policy" "test" {
  enabled     = false
  priority    = %d
  action      = []
  rulesets    = []
  content     = []
  new_action  = "alert"
  description = %q
}
`, priority, description)
}

func testAccIDSPolicyRuleConfig(message string) string {
	return fmt.Sprintf(`
resource "opnsense_ids_policy_rule" "test" {
  sid     = "9000001"
  enabled = false
  action  = "alert"
  message = %q
  source  = "acceptance"
}
`, message)
}

func testAccIDSUserRuleConfig(description string) string {
	return fmt.Sprintf(`
resource "opnsense_ids_user_rule" "test" {
  enabled     = false
  source      = "192.0.2.1"
  destination = "any"
  fingerprint = "BB:BB:BB:BB:BB:BB:BB:BB:BB:BB:BB:BB:BB:BB:BB:BB:BB:BB:BB:BB"
  description = %q
  action      = "alert"
  bypass      = false
}
`, description)
}
