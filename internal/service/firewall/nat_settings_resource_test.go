package firewall_test

import (
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFirewallNATSettingsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallNATSettingsResourceConfig("hybrid"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_firewall_nat_settings.test", "mode", "hybrid"),
					resource.TestCheckResourceAttr("opnsense_firewall_nat_settings.test", "id", "settings"),
				),
			},
			{
				Config: testAccFirewallNATSettingsResourceConfig("automatic"),
				Check: resource.TestCheckResourceAttr(
					"opnsense_firewall_nat_settings.test", "mode", "automatic",
				),
			},
			{
				ResourceName:      "opnsense_firewall_nat_settings.test",
				ImportState:       true,
				ImportStateId:     "settings",
				ImportStateVerify: true,
			},
		},
	})
}

func testAccFirewallNATSettingsResourceConfig(mode string) string {
	return `
resource "opnsense_firewall_nat_settings" "test" {
  mode = "` + mode + `"
}
`
}
