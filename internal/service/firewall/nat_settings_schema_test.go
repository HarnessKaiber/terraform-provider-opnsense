package firewall

import (
	"testing"

	"github.com/browningluke/opnsense-go/pkg/api"
	clientfirewall "github.com/browningluke/opnsense-go/pkg/firewall"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestNATSettingsConversions(t *testing.T) {
	model := &natSettingsResourceModel{Mode: types.StringValue("hybrid")}
	apiSettings := natSettingsToAPI(model)
	assert.Equal(t, api.SelectedMap("hybrid"), apiSettings.General.SNATMode)

	converted := natSettingsFromAPI(&clientfirewall.NATSettings{
		General: clientfirewall.NATGeneral{SNATMode: api.SelectedMap("advanced")},
	})
	assert.Equal(t, "advanced", converted.Mode.ValueString())
	assert.Equal(t, "settings", converted.ID.ValueString())
}
