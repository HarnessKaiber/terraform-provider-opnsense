package ids_test

import (
	"context"
	"os"
	"testing"

	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
)

// TestAccIDSSettingsGet verifies that the provider's API user can read the
// native Suricata/IDS configuration. It is deliberately read-only.
func TestAccIDSSettingsGet(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("acceptance test; set TF_ACC=1 to run")
	}
	acctest.AccPreCheck(t)

	client := opnsense.NewClient(api.NewClient(api.Options{
		Uri:       os.Getenv("OPNSENSE_URI"),
		APIKey:    os.Getenv("OPNSENSE_API_KEY"),
		APISecret: os.Getenv("OPNSENSE_API_SECRET"),
	}))

	settings, err := client.Ids().SettingsGet(context.Background())
	if err != nil {
		t.Fatalf("read IDS settings: %s", err)
	}
	if settings == nil {
		t.Fatal("read IDS settings returned no settings")
	}
	if settings.Ids.General.Enabled == "" {
		t.Fatal("read IDS settings returned an empty IDS general configuration")
	}
}
