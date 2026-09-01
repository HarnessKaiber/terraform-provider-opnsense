package zenarmor

import (
	"context"
	"fmt"

	"github.com/browningluke/opnsense-go/pkg/api"
	clientcore "github.com/browningluke/opnsense-go/pkg/core"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
	clientzenarmor "github.com/browningluke/opnsense-go/pkg/zenarmor"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &statusDataSource{}
var _ datasource.DataSourceWithConfigure = &statusDataSource{}

type statusDataSource struct{ client opnsense.Client }

func newStatusDataSource() datasource.DataSource { return &statusDataSource{} }

func (d *statusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zenarmor_status"
}

func (d *statusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = statusDataSourceSchema()
}

func (d *statusDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	apiClient, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *api.Client, got %T.", req.ProviderData))
		return
	}
	d.client = opnsense.NewClient(apiClient)
}

func (d *statusDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	info, err := d.client.Core().FirmwareInfo(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to query Zenarmor installation", fmt.Sprintf("Unable to query the OPNsense firmware inventory: %s", err))
		return
	}
	model := statusModelFromPackages(info.Plugin)
	if model.Installed.ValueBool() {
		status, err := d.client.Zenarmor().Status(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Unable to query Zenarmor status", err.Error())
			return
		}
		model = statusModelFromAPI(model, status)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func statusModelFromPackages(packages []clientcore.Package) statusDataSourceModel {
	model := statusDataSourceModel{
		ID:             types.StringValue("os-sensei"),
		Installed:      types.BoolValue(false),
		PluginVersion:  types.StringNull(),
		AgentVersion:   types.StringNull(),
		UpdaterVersion: types.StringNull(),
		EngineVersion:  types.StringNull(), EngineStatus: types.BoolNull(),
		ApplicationDatabaseVersion: types.StringNull(), ThreatDatabaseVersion: types.StringNull(),
		Edition: types.StringNull(), LicenseStatus: types.StringNull(), SupportedFeatures: []types.String{},
		TLSInspectionSupported: types.BoolValue(false), FullTLSInspectionSupported: types.BoolValue(false), CloudAccessSupported: types.BoolValue(false),
	}
	for _, pkg := range packages {
		if pkg.Installed != "1" {
			continue
		}
		switch pkg.Name {
		case "os-sensei":
			model.Installed = types.BoolValue(true)
			model.PluginVersion = types.StringValue(pkg.Version)
		case "os-sensei-agent":
			model.AgentVersion = types.StringValue(pkg.Version)
		case "os-sensei-updater":
			model.UpdaterVersion = types.StringValue(pkg.Version)
		}
	}
	return model
}

func statusModelFromAPI(model statusDataSourceModel, status *clientzenarmor.Status) statusDataSourceModel {
	model.EngineVersion = nullableString(status.Engine.Version)
	model.EngineStatus = types.BoolValue(status.Eastpect.Status)
	model.ApplicationDatabaseVersion = nullableString(status.DatabaseVersion.Version)
	if model.ApplicationDatabaseVersion.IsNull() {
		model.ApplicationDatabaseVersion = nullableString(status.Database.Info.Version)
	}
	model.Edition = nullableString(status.License)
	model.LicenseStatus = nullableString(status.License)
	model.CloudAccessSupported = types.BoolValue(status.CloudThreatIntel)
	features := make([]types.String, 0, 1)
	if status.CloudThreatIntel {
		features = append(features, types.StringValue("cloud_threat_intelligence"))
	}
	model.SupportedFeatures = features
	return model
}

func nullableString(value string) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}
