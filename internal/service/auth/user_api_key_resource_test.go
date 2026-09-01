package auth_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserAPIKeyResource(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("acceptance test; set TF_ACC=1 to run")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserAPIKeyResourceConfig(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_auth_user_api_key.test", "id"),
					resource.TestCheckResourceAttrSet("opnsense_auth_user_api_key.test", "key"),
					resource.TestCheckResourceAttrSet("opnsense_auth_user_api_key.test", "secret"),
				),
			},
			{
				ResourceName:      "opnsense_auth_user_api_key.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"secret",
				},
			},
		},
	})
}

func testAccUserAPIKeyResourceConfig(t *testing.T) string {
	t.Helper()

	client := opnsense.NewClient(api.NewClient(api.Options{
		Uri:           os.Getenv("OPNSENSE_URI"),
		APIKey:        os.Getenv("OPNSENSE_API_KEY"),
		APISecret:     os.Getenv("OPNSENSE_API_SECRET"),
		AllowInsecure: os.Getenv("OPNSENSE_ALLOW_INSECURE") == "true",
	}))
	keys, err := client.Auth().UserGetAllApiKeys(context.Background())
	if err != nil {
		t.Fatalf("unable to determine the API credential owner: %s", err)
	}

	for _, key := range keys.Rows {
		if key.Key == os.Getenv("OPNSENSE_API_KEY") {
			return fmt.Sprintf(`
resource "opnsense_auth_user_api_key" "test" {
  username = %q
}
`, key.Username)
		}
	}
	t.Fatal("the configured OPNSENSE_API_KEY was not found in the OPNsense API-key list")
	return ""
}
