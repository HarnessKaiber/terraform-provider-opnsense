package ids

import (
	"context"

	"github.com/browningluke/opnsense-go/pkg/opnsense"
)

// reloadRulesIfEnabled avoids OPNsense returning an error from reload_rules
// when IDS is intentionally disabled. A later enable/reconfigure applies all
// stored policy, ruleset, and rule changes together.
func reloadRulesIfEnabled(ctx context.Context, client opnsense.Client) error {
	settings, err := client.Ids().SettingsGet(ctx)
	if err != nil || settings.Ids.General.Enabled != "1" {
		return err
	}

	return client.Ids().Client().ReconfigureService(ctx, "/ids/service/reload_rules")
}
