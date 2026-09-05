package firewall

import (
	"context"
	"fmt"

	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &natSettingsResource{}
var _ resource.ResourceWithConfigure = &natSettingsResource{}
var _ resource.ResourceWithImportState = &natSettingsResource{}

type natSettingsResource struct {
	client opnsense.Client
}

func newNATSettingsResource() resource.Resource {
	return &natSettingsResource{}
}

func (r *natSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_nat_settings"
}

func (r *natSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = natSettingsResourceSchema()
}

func (r *natSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *api.Client, got %T.", req.ProviderData),
		)
		return
	}

	r.client = opnsense.NewClient(client)
}

func (r *natSettingsResource) read(ctx context.Context) (*natSettingsResourceModel, error) {
	settings, err := r.client.Firewall().NATSettingsGet(ctx)
	if err != nil {
		return nil, err
	}
	return natSettingsFromAPI(&settings.Filter), nil
}

func (r *natSettingsResource) apply(ctx context.Context, data *natSettingsResourceModel) error {
	if _, err := r.client.Firewall().NATSettingsUpdate(ctx, natSettingsToAPI(data)); err != nil {
		return err
	}
	_, err := r.client.Firewall().NATSettingsApply(ctx)
	return err
}

func (r *natSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data natSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.apply(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update outbound NAT settings: %s", err))
		return
	}

	updated, err := r.read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated outbound NAT settings: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, updated)...)
}

func (r *natSettingsResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	data, err := r.read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read outbound NAT settings: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *natSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data natSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.apply(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update outbound NAT settings: %s", err))
		return
	}

	updated, err := r.read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read updated outbound NAT settings: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, updated)...)
}

func (r *natSettingsResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	data := natSettingsResourceModel{
		Mode: types.StringValue("automatic"),
		ID:   types.StringValue("settings"),
	}
	if err := r.apply(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to restore automatic outbound NAT settings: %s", err))
	}
}

func (r *natSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "settings" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Outbound NAT settings must be imported with the singleton ID 'settings'; got %q.", req.ID),
		)
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
