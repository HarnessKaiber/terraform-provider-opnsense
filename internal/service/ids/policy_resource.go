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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &policyResource{}
var _ resource.ResourceWithConfigure = &policyResource{}
var _ resource.ResourceWithImportState = &policyResource{}

type policyResource struct{ client opnsense.Client }
type policyResourceModel struct {
	Enabled     types.Bool   `tfsdk:"enabled"`
	Priority    types.Int64  `tfsdk:"priority"`
	Action      types.Set    `tfsdk:"action"`
	Rulesets    types.Set    `tfsdk:"rulesets"`
	Content     types.Set    `tfsdk:"content"`
	NewAction   types.String `tfsdk:"new_action"`
	Description types.String `tfsdk:"description"`
	ID          types.String `tfsdk:"id"`
}

func newPolicyResource() resource.Resource { return &policyResource{} }
func (r *policyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ids_policy"
}
func (r *policyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Version: 1, MarkdownDescription: "A Suricata intrusion-detection policy.", Attributes: map[string]schema.Attribute{
		"enabled": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true)}, "priority": schema.Int64Attribute{Optional: true, Computed: true, Default: int64default.StaticInt64(1)}, "action": schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType, Default: setdefault.StaticValue(tools.EmptySetValue(types.StringType))}, "rulesets": schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType, Default: setdefault.StaticValue(tools.EmptySetValue(types.StringType))}, "content": schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType, Default: setdefault.StaticValue(tools.EmptySetValue(types.StringType))}, "new_action": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("")}, "description": schema.StringAttribute{Optional: true}, "id": schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}}}
}
func (r *policyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func policyToAPI(d *policyResourceModel) (*clientids.Policy, error) {
	var action, rulesets, content []string
	for _, v := range []struct {
		s types.Set
		d *[]string
	}{{d.Action, &action}, {d.Rulesets, &rulesets}, {d.Content, &content}} {
		if diags := v.s.ElementsAs(context.Background(), v.d, false); diags.HasError() {
			return nil, fmt.Errorf("invalid IDS policy set")
		}
	}
	return &clientids.Policy{Enabled: tools.BoolToString(d.Enabled.ValueBool()), Priority: fmt.Sprint(d.Priority.ValueInt64()), Action: action, Rulesets: rulesets, Content: content, NewAction: api.SelectedMap(d.NewAction.ValueString()), Description: d.Description.ValueString()}, nil
}
func policyFromAPI(p *clientids.Policy, id types.String) *policyResourceModel {
	return &policyResourceModel{Enabled: types.BoolValue(tools.StringToBool(p.Enabled)), Priority: types.Int64Value(int64(tools.StringToFloat64(p.Priority))), Action: tools.StringSliceToSet(p.Action), Rulesets: tools.StringSliceToSet(p.Rulesets), Content: tools.StringSliceToSet(p.Content), NewAction: types.StringValue(p.NewAction.String()), Description: tools.StringOrNull(p.Description), ID: id}
}
func (r *policyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var d policyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &d)...)
	if resp.Diagnostics.HasError() {
		return
	}
	p, e := policyToAPI(&d)
	if e != nil {
		resp.Diagnostics.AddError("Client Error", e.Error())
		return
	}
	id, e := r.client.Ids().AddPolicy(ctx, p)
	if e != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create IDS policy: %s", e))
		return
	}
	if e = reloadRulesIfEnabled(ctx, r.client); e != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to reload IDS rules: %s", e))
		return
	}
	d.ID = types.StringValue(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &d)...)
}
func (r *policyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var d policyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &d)...)
	if resp.Diagnostics.HasError() {
		return
	}
	p, e := r.client.Ids().GetPolicy(ctx, d.ID.ValueString())
	if e != nil {
		var notFound *errs.NotFoundError
		if errors.As(e, &notFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read IDS policy: %s", e))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, policyFromAPI(p, d.ID))...)
}
func (r *policyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var d policyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &d)...)
	if resp.Diagnostics.HasError() {
		return
	}
	p, e := policyToAPI(&d)
	if e != nil {
		resp.Diagnostics.AddError("Client Error", e.Error())
		return
	}
	if e = r.client.Ids().UpdatePolicy(ctx, d.ID.ValueString(), p); e != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update IDS policy: %s", e))
		return
	}
	if e = reloadRulesIfEnabled(ctx, r.client); e != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to reload IDS rules: %s", e))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &d)...)
}
func (r *policyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var d policyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &d)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if e := r.client.Ids().DeletePolicy(ctx, d.ID.ValueString()); e != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete IDS policy: %s", e))
		return
	}
	if e := reloadRulesIfEnabled(ctx, r.client); e != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to reload IDS rules: %s", e))
	}
}
func (r *policyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
