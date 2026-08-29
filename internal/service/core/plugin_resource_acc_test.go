package core_test

import (
	"testing"

	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPluginResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `resource "opnsense_plugin" "test" { name = "os-debug" }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_plugin.test", "name", "os-debug"),
					resource.TestCheckResourceAttrSet("opnsense_plugin.test", "version"),
				),
			},
			{ResourceName: "opnsense_plugin.test", ImportState: true, ImportStateVerify: true},
		},
	})
}
