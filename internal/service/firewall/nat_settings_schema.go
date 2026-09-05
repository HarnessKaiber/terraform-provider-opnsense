package firewall

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	clientfirewall "github.com/browningluke/opnsense-go/pkg/firewall"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type natSettingsResourceModel struct {
	Mode types.String `tfsdk:"mode"`
	ID   types.String `tfsdk:"id"`
}

func natSettingsResourceSchema() schema.Schema {
	return schema.Schema{
		Version:             1,
		MarkdownDescription: "Singleton OPNsense outbound source NAT settings.",
		Attributes: map[string]schema.Attribute{
			"mode": schema.StringAttribute{
				MarkdownDescription: "Outbound source NAT rule generation mode. Must be `automatic`, `hybrid`, `advanced`, or `disabled`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("automatic", "hybrid", "advanced", "disabled"),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Singleton resource identifier.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func natSettingsToAPI(data *natSettingsResourceModel) *clientfirewall.NATSettings {
	return &clientfirewall.NATSettings{
		General: clientfirewall.NATGeneral{
			SNATMode: api.SelectedMap(data.Mode.ValueString()),
		},
	}
}

func natSettingsFromAPI(settings *clientfirewall.NATSettings) *natSettingsResourceModel {
	return &natSettingsResourceModel{
		Mode: types.StringValue(settings.General.SNATMode.String()),
		ID:   types.StringValue("settings"),
	}
}
