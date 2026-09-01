package ids

import (
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type settingsResourceModel struct {
	Enabled      types.Bool   `tfsdk:"enabled"`
	Mode         types.String `tfsdk:"mode"`
	Interfaces   types.Set    `tfsdk:"interfaces"`
	HomeNetworks types.Set    `tfsdk:"home_networks"`
	Promiscuous  types.Bool   `tfsdk:"promiscuous"`
	ID           types.String `tfsdk:"id"`
}

func settingsResourceSchema() schema.Schema {
	return schema.Schema{
		Version:             1,
		MarkdownDescription: "Singleton OPNsense Suricata IDS settings.",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether to enable intrusion detection. Defaults to `false`.",
				Optional:            true, Computed: true, Default: booldefault.StaticBool(false),
			},
			"mode": schema.StringAttribute{
				MarkdownDescription: "Packet capture mode for intrusion detection. Defaults to `pcap`.",
				Optional:            true, Computed: true, Default: stringdefault.StaticString("pcap"),
			},
			"interfaces": schema.SetAttribute{
				MarkdownDescription: "Interfaces on which to inspect traffic. Defaults to an empty set.",
				Optional:            true, Computed: true, ElementType: types.StringType, Default: setdefault.StaticValue(tools.EmptySetValue(types.StringType)),
			},
			"home_networks": schema.SetAttribute{
				MarkdownDescription: "Networks treated as local by Suricata. Defaults to an empty set.",
				Optional:            true, Computed: true, ElementType: types.StringType, Default: setdefault.StaticValue(tools.EmptySetValue(types.StringType)),
			},
			"promiscuous": schema.BoolAttribute{
				MarkdownDescription: "Whether to enable promiscuous mode. Defaults to `false`.",
				Optional:            true, Computed: true, Default: booldefault.StaticBool(false),
			},
			"id": schema.StringAttribute{MarkdownDescription: "Singleton resource identifier.", Computed: true},
		},
	}
}
