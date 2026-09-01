package zenarmor

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type statusDataSourceModel struct {
	ID                         types.String   `tfsdk:"id"`
	Installed                  types.Bool     `tfsdk:"installed"`
	PluginVersion              types.String   `tfsdk:"plugin_version"`
	AgentVersion               types.String   `tfsdk:"agent_version"`
	UpdaterVersion             types.String   `tfsdk:"updater_version"`
	EngineVersion              types.String   `tfsdk:"engine_version"`
	EngineStatus               types.Bool     `tfsdk:"engine_status"`
	ApplicationDatabaseVersion types.String   `tfsdk:"application_database_version"`
	ThreatDatabaseVersion      types.String   `tfsdk:"threat_database_version"`
	Edition                    types.String   `tfsdk:"edition"`
	LicenseStatus              types.String   `tfsdk:"license_status"`
	SupportedFeatures          []types.String `tfsdk:"supported_features"`
	TLSInspectionSupported     types.Bool     `tfsdk:"tls_inspection_supported"`
	FullTLSInspectionSupported types.Bool     `tfsdk:"full_tls_inspection_supported"`
	CloudAccessSupported       types.Bool     `tfsdk:"cloud_access_supported"`
}

func statusDataSourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Reports package, engine, database, licence, and capability information exposed by the local OPNsense Zenarmor controller.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Stable identifier for the local Zenarmor installation.",
				Computed:            true,
			},
			"installed": schema.BoolAttribute{
				MarkdownDescription: "Whether the `os-sensei` Zenarmor plugin is installed.",
				Computed:            true,
			},
			"plugin_version": schema.StringAttribute{
				MarkdownDescription: "Installed `os-sensei` package version, or null when it is not installed.",
				Computed:            true,
			},
			"agent_version": schema.StringAttribute{
				MarkdownDescription: "Installed `os-sensei-agent` package version, or null when it is not installed.",
				Computed:            true,
			},
			"updater_version": schema.StringAttribute{
				MarkdownDescription: "Installed `os-sensei-updater` package version, or null when it is not installed.",
				Computed:            true,
			},
			"engine_version":                schema.StringAttribute{Computed: true, MarkdownDescription: "Zenarmor engine version."},
			"engine_status":                 schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the Zenarmor engine is running."},
			"application_database_version":  schema.StringAttribute{Computed: true, MarkdownDescription: "Zenarmor application database version."},
			"threat_database_version":       schema.StringAttribute{Computed: true, MarkdownDescription: "Threat database version when separately advertised by the installed release."},
			"edition":                       schema.StringAttribute{Computed: true, MarkdownDescription: "Zenarmor edition reported by the local controller."},
			"license_status":                schema.StringAttribute{Computed: true, MarkdownDescription: "Raw licence status reported by the local controller."},
			"supported_features":            schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Deterministically ordered capabilities positively advertised by the local controller."},
			"tls_inspection_supported":      schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether TLS inspection is positively advertised as supported."},
			"full_tls_inspection_supported": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether full TLS inspection is positively advertised as supported."},
			"cloud_access_supported":        schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether cloud threat-intelligence access is enabled."},
		},
	}
}
