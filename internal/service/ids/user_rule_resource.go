package ids

import (
	"context"
	"errors"
	"fmt"

	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/errs"
	clientids "github.com/browningluke/opnsense-go/pkg/ids"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &userRuleResource{}
var _ resource.ResourceWithConfigure = &userRuleResource{}
var _ resource.ResourceWithImportState = &userRuleResource{}

type userRuleResource struct{ client opnsense.Client }
type userRuleResourceModel struct {
	Enabled     types.Bool   `tfsdk:"enabled"`
	Source      types.String `tfsdk:"source"`
	Destination types.String `tfsdk:"destination"`
	Fingerprint types.String `tfsdk:"fingerprint"`
	Description types.String `tfsdk:"description"`
	Action      types.String `tfsdk:"action"`
	Bypass      types.Bool   `tfsdk:"bypass"`
	ID          types.String `tfsdk:"id"`
}

func newUserRuleResource() resource.Resource { return &userRuleResource{} }
func (r *userRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ids_user_rule"
}
func (r *userRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Version: 1, MarkdownDescription: "A custom Suricata IDS rule.", Attributes: map[string]schema.Attribute{
		"enabled": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)}, "source": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("any")}, "destination": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("any")}, "fingerprint": schema.StringAttribute{Required: true}, "description": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")}, "action": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("alert")}, "bypass": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)}, "id": schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}}}
}
func (r *userRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func userRuleToAPI(d *userRuleResourceModel) *clientids.UserRule {
	return &clientids.UserRule{Enabled: tools.BoolToString(d.Enabled.ValueBool()), Source: d.Source.ValueString(), Destination: d.Destination.ValueString(), Fingerprint: d.Fingerprint.ValueString(), Description: d.Description.ValueString(), Action: api.SelectedMap(d.Action.ValueString()), Bypass: tools.BoolToString(d.Bypass.ValueBool())}
}
func userRuleFromAPI(p *clientids.UserRule, id types.String) *userRuleResourceModel {
	return &userRuleResourceModel{Enabled: types.BoolValue(tools.StringToBool(p.Enabled)), Source: types.StringValue(p.Source), Destination: types.StringValue(p.Destination), Fingerprint: types.StringValue(p.Fingerprint), Description: types.StringValue(p.Description), Action: types.StringValue(p.Action.String()), Bypass: types.BoolValue(tools.StringToBool(p.Bypass)), ID: id}
}
func (r *userRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var d userRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &d)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := r.client.Ids().AddUserRule(ctx, userRuleToAPI(&d))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create IDS user rule: %s", err))
		return
	}
	if err = reloadRulesIfEnabled(ctx, r.client); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to reload IDS rules: %s", err))
		return
	}
	d.ID = types.StringValue(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &d)...)
}
func (r *userRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var d userRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &d)...)
	if resp.Diagnostics.HasError() {
		return
	}
	p, err := r.client.Ids().GetUserRule(ctx, d.ID.ValueString())
	if err != nil {
		var notFound *errs.NotFoundError
		if errors.As(err, &notFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read IDS user rule: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, userRuleFromAPI(p, d.ID))...)
}
func (r *userRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var d userRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &d)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Ids().UpdateUserRule(ctx, d.ID.ValueString(), userRuleToAPI(&d)); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update IDS user rule: %s", err))
		return
	}
	if err := reloadRulesIfEnabled(ctx, r.client); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to reload IDS rules: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &d)...)
}
func (r *userRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var d userRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &d)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Ids().DeleteUserRule(ctx, d.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete IDS user rule: %s", err))
		return
	}
	if err := reloadRulesIfEnabled(ctx, r.client); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to reload IDS rules: %s", err))
	}
}
func (r *userRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
