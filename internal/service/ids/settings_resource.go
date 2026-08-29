package ids

import (
	"context"
	"fmt"

	"github.com/browningluke/opnsense-go/pkg/api"
	clientids "github.com/browningluke/opnsense-go/pkg/ids"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &settingsResource{}
var _ resource.ResourceWithConfigure = &settingsResource{}
var _ resource.ResourceWithImportState = &settingsResource{}

type settingsResource struct{ client opnsense.Client }

func newSettingsResource() resource.Resource { return &settingsResource{} }
func (r *settingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ids_settings"
}
func (r *settingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = settingsResourceSchema()
}
func (r *settingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *api.Client, got %T.", req.ProviderData))
		return
	}
	r.client = opnsense.NewClient(c)
}
func settingsToAPI(d *settingsResourceModel) (*clientids.Settings, error) {
	var interfaces, homeNetworks []string
	for _, v := range []struct {
		s types.Set
		d *[]string
	}{{d.Interfaces, &interfaces}, {d.HomeNetworks, &homeNetworks}} {
		if diags := v.s.ElementsAs(context.Background(), v.d, false); diags.HasError() {
			return nil, fmt.Errorf("invalid IDS settings set")
		}
	}
	return &clientids.Settings{General: clientids.General{Enabled: tools.BoolToString(d.Enabled.ValueBool()), Mode: api.SelectedMap(d.Mode.ValueString()), Interfaces: interfaces, HomeNet: homeNetworks, Promisc: tools.BoolToString(d.Promiscuous.ValueBool())}}, nil
}
func settingsFromAPI(s *clientids.Settings) *settingsResourceModel {
	return &settingsResourceModel{Enabled: types.BoolValue(tools.StringToBool(s.General.Enabled)), Mode: types.StringValue(s.General.Mode.String()), Interfaces: tools.StringSliceToSet(s.General.Interfaces), HomeNetworks: tools.StringSliceToSet(s.General.HomeNet), Promiscuous: types.BoolValue(tools.StringToBool(s.General.Promisc)), ID: types.StringValue("settings")}
}
func (r *settingsResource) read(ctx context.Context) (*settingsResourceModel, error) {
	settings, err := r.client.Ids().SettingsGet(ctx)
	if err != nil {
		return nil, err
	}
	return settingsFromAPI(&settings.Ids), nil
}
func (r *settingsResource) apply(ctx context.Context, d *settingsResourceModel) error {
	settings, err := settingsToAPI(d)
	if err != nil {
		return err
	}
	if _, err = r.client.Ids().SettingsUpdate(ctx, settings); err != nil {
		return err
	}
	return r.client.Ids().Client().ReconfigureService(ctx, "/ids/service/reconfigure")
}
func (r *settingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var d settingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &d)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &d); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update IDS settings: %s", err))
		return
	}
	d.ID = types.StringValue("settings")
	resp.Diagnostics.Append(resp.State.Set(ctx, &d)...)
}
func (r *settingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	d, err := r.read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read IDS settings: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, d)...)
}
func (r *settingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var d settingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &d)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, &d); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update IDS settings: %s", err))
		return
	}
	d.ID = types.StringValue("settings")
	resp.Diagnostics.Append(resp.State.Set(ctx, &d)...)
}
func (r *settingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	current, err := r.read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read IDS settings before disabling IDS: %s", err))
		return
	}
	current.Enabled = types.BoolValue(false)
	if err := r.apply(ctx, current); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to disable IDS settings: %s", err))
	}
}
func (r *settingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.AddError("Import not supported", "IDS settings are a singleton. Declare the resource with its desired configuration instead of importing it.")
}
