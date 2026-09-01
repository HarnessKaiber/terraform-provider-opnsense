package auth

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type userAPIKeyResourceModel struct {
	Username types.String `tfsdk:"username"`
	Key      types.String `tfsdk:"key"`
	Secret   types.String `tfsdk:"secret"`
	ID       types.String `tfsdk:"id"`
}

func userAPIKeyResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "API keys authenticate machine-to-machine access to OPNsense as a local user. The secret is returned by OPNsense only when the key is created, so Terraform stores it as sensitive state.",
		Version:             1,
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				MarkdownDescription: "Name of the local OPNsense user that owns this API key.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "API key identifier returned by OPNsense.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"secret": schema.StringAttribute{
				MarkdownDescription: "API key secret, returned only once when the key is created.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Encoded API-key identifier used by OPNsense to revoke the key.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}
