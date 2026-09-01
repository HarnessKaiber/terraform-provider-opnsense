package ids

import (
	"context"
	"fmt"

	"github.com/browningluke/opnsense-go/pkg/api"
	clientids "github.com/browningluke/opnsense-go/pkg/ids"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &rulesetResource{}
var _ resource.ResourceWithConfigure = &rulesetResource{}
var _ resource.ResourceWithImportState = &rulesetResource{}

type rulesetResource struct{ client opnsense.Client }
type rulesetResourceModel struct {
	Filename    types.String `tfsdk:"filename"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Description types.String `tfsdk:"description"`
	ID          types.String `tfsdk:"id"`
}

func newRulesetResource() resource.Resource { return &rulesetResource{} }
func (r *rulesetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ids_ruleset"
}
func (r *rulesetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Version: 1, MarkdownDescription: "An installable Suricata ruleset. Removing this resource disables, but does not delete, the OPNsense-provided ruleset.", Attributes: map[string]schema.Attribute{
		"filename": schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}}, "enabled": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false)}, "description": schema.StringAttribute{Computed: true}, "id": schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}}}
}
func (r *rulesetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func rulesetFromAPI(v *clientids.Ruleset) *rulesetResourceModel {
	return &rulesetResourceModel{Filename: types.StringValue(v.Filename), Enabled: types.BoolValue(tools.StringToBool(v.Enabled)), Description: types.StringValue(v.Description), ID: types.StringValue(v.Filename)}
}
func (r *rulesetResource) read(ctx context.Context, filename string) (*rulesetResourceModel, error) {
	v, err := r.client.Ids().RulesetGet(ctx, filename)
	if err != nil {
		return nil, err
	}
	if v.Filename == "" {
		return nil, fmt.Errorf("IDS ruleset %q was not found", filename)
	}
	return rulesetFromAPI(v), nil
}
func (r *rulesetResource) apply(ctx context.Context, filename string, enabled bool) error {
	result, err := r.client.Ids().RulesetToggle(ctx, filename, tools.BoolToString(enabled))
	if err != nil {
		return err
	}
	if result.Status != "0" && result.Status != "1" {
		return fmt.Errorf("unexpected ruleset toggle status %q", result.Status)
	}
	return reloadRulesIfEnabled(ctx, r.client)
}
func (r *rulesetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var d rulesetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &d)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, d.Filename.ValueString(), d.Enabled.ValueBool()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update IDS ruleset: %s", err))
		return
	}
	current, err := r.read(ctx, d.Filename.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read IDS ruleset: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
}
func (r *rulesetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var d rulesetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &d)...)
	if resp.Diagnostics.HasError() {
		return
	}
	current, err := r.read(ctx, d.Filename.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read IDS ruleset: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
}
func (r *rulesetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var d rulesetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &d)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, d.Filename.ValueString(), d.Enabled.ValueBool()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update IDS ruleset: %s", err))
		return
	}
	current, err := r.read(ctx, d.Filename.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read IDS ruleset: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, current)...)
}
func (r *rulesetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var d rulesetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &d)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, d.Filename.ValueString(), false); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to disable IDS ruleset: %s", err))
	}
}
func (r *rulesetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("filename"), req, resp)
}
