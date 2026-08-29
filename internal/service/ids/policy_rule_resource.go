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

var _ resource.Resource = &policyRuleResource{}
var _ resource.ResourceWithConfigure = &policyRuleResource{}
var _ resource.ResourceWithImportState = &policyRuleResource{}

type policyRuleResource struct{ client opnsense.Client }
type policyRuleResourceModel struct {
	SID     types.String `tfsdk:"sid"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Action  types.String `tfsdk:"action"`
	Message types.String `tfsdk:"message"`
	Source  types.String `tfsdk:"source"`
	ID      types.String `tfsdk:"id"`
}

func newPolicyRuleResource() resource.Resource { return &policyRuleResource{} }
func (r *policyRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ids_policy_rule"
}
func (r *policyRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Version: 1, MarkdownDescription: "A Suricata policy-rule override.", Attributes: map[string]schema.Attribute{
		"sid": schema.StringAttribute{Required: true}, "enabled": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)}, "action": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("alert")}, "message": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")}, "source": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")}, "id": schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}}}
}
func (r *policyRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func policyRuleToAPI(d *policyRuleResourceModel) *clientids.PolicyRule {
	return &clientids.PolicyRule{SID: d.SID.ValueString(), Enabled: tools.BoolToString(d.Enabled.ValueBool()), Action: api.SelectedMap(d.Action.ValueString()), Message: d.Message.ValueString(), Source: d.Source.ValueString()}
}
func policyRuleFromAPI(p *clientids.PolicyRule, id types.String) *policyRuleResourceModel {
	return &policyRuleResourceModel{SID: types.StringValue(p.SID), Enabled: types.BoolValue(tools.StringToBool(p.Enabled)), Action: types.StringValue(p.Action.String()), Message: types.StringValue(p.Message), Source: types.StringValue(p.Source), ID: id}
}
func (r *policyRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var d policyRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &d)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := r.client.Ids().AddPolicyRule(ctx, policyRuleToAPI(&d))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create IDS policy rule: %s", err))
		return
	}
	if err = reloadRulesIfEnabled(ctx, r.client); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to reload IDS rules: %s", err))
		return
	}
	d.ID = types.StringValue(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &d)...)
}
func (r *policyRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var d policyRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &d)...)
	if resp.Diagnostics.HasError() {
		return
	}
	p, err := r.client.Ids().GetPolicyRule(ctx, d.ID.ValueString())
	if err != nil {
		var notFound *errs.NotFoundError
		if errors.As(err, &notFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read IDS policy rule: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, policyRuleFromAPI(p, d.ID))...)
}
func (r *policyRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var d policyRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &d)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Ids().UpdatePolicyRule(ctx, d.ID.ValueString(), policyRuleToAPI(&d)); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update IDS policy rule: %s", err))
		return
	}
	if err := reloadRulesIfEnabled(ctx, r.client); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to reload IDS rules: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &d)...)
}
func (r *policyRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var d policyRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &d)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Ids().DeletePolicyRule(ctx, d.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete IDS policy rule: %s", err))
		return
	}
	if err := reloadRulesIfEnabled(ctx, r.client); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to reload IDS rules: %s", err))
	}
}
func (r *policyRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
