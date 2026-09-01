package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/browningluke/opnsense-go/pkg/api"
	opnauth "github.com/browningluke/opnsense-go/pkg/auth"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type privilegeDataModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Match  types.String `tfsdk:"match"`
	Users  types.List   `tfsdk:"users"`
	Groups types.List   `tfsdk:"groups"`
}
type privilegeAssignmentModel struct {
	PrivilegeID types.String `tfsdk:"privilege_id"`
	PrincipalID types.String `tfsdk:"principal_id"`
	ID          types.String `tfsdk:"id"`
}
type privilegeDataSource struct{ client opnsense.Client }
type userPrivilegeResource struct{ client opnsense.Client }
type groupPrivilegeResource struct{ client opnsense.Client }

func newPrivilegeDataSource() datasource.DataSource { return &privilegeDataSource{} }
func newUserPrivilegeResource() resource.Resource   { return &userPrivilegeResource{} }
func newGroupPrivilegeResource() resource.Resource  { return &groupPrivilegeResource{} }

func privilegeDataSourceSchema() dschema.Schema {
	return dschema.Schema{MarkdownDescription: "Looks up an OPNsense privilege by its ID.", Attributes: map[string]dschema.Attribute{
		"id": dschema.StringAttribute{MarkdownDescription: "Privilege ID.", Required: true}, "name": dschema.StringAttribute{MarkdownDescription: "Privilege name.", Computed: true}, "match": dschema.StringAttribute{MarkdownDescription: "API and GUI paths granted by the privilege.", Computed: true}, "users": dschema.ListAttribute{MarkdownDescription: "User IDs assigned directly.", Computed: true, ElementType: types.StringType}, "groups": dschema.ListAttribute{MarkdownDescription: "Group IDs assigned directly.", Computed: true, ElementType: types.StringType},
	}}
}
func privilegeAssignmentSchema(kind string) schema.Schema {
	return schema.Schema{Version: 1, MarkdownDescription: "Assigns an OPNsense privilege directly to a " + kind + " while preserving other assignments.", Attributes: map[string]schema.Attribute{
		"privilege_id": schema.StringAttribute{MarkdownDescription: "OPNsense privilege ID.", Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"principal_id": schema.StringAttribute{MarkdownDescription: "UUID of the " + kind + ".", Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"id":           schema.StringAttribute{MarkdownDescription: "Composite assignment identifier.", Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}
func configureData(c *opnsense.Client, v any, errf func(string, string)) {
	a, ok := v.(*api.Client)
	if !ok {
		errf("Unexpected Configure Type", fmt.Sprintf("Expected *api.Client, got: %T", v))
		return
	}
	*c = opnsense.NewClient(a)
}
func (d *privilegeDataSource) Metadata(ctx context.Context, q datasource.MetadataRequest, p *datasource.MetadataResponse) {
	p.TypeName = q.ProviderTypeName + "_auth_privilege"
}
func (d *privilegeDataSource) Schema(ctx context.Context, q datasource.SchemaRequest, p *datasource.SchemaResponse) {
	p.Schema = privilegeDataSourceSchema()
}
func (d *privilegeDataSource) Configure(ctx context.Context, q datasource.ConfigureRequest, p *datasource.ConfigureResponse) {
	if q.ProviderData != nil {
		configureData(&d.client, q.ProviderData, p.Diagnostics.AddError)
	}
}
func (d *privilegeDataSource) Read(ctx context.Context, q datasource.ReadRequest, p *datasource.ReadResponse) {
	var m privilegeDataModel
	p.Diagnostics.Append(q.Config.Get(ctx, &m)...)
	all, e := d.client.Auth().PrivilegeGetAll(ctx)
	if e != nil {
		p.Diagnostics.AddError("Client Error", e.Error())
		return
	}
	for _, x := range all.Rows {
		if x.Id == m.ID.ValueString() {
			m.Name = types.StringValue(x.Name)
			m.Match = types.StringValue(x.Match)
			m.Users = stringList(x.Users)
			m.Groups = stringList(x.Groups)
			p.Diagnostics.Append(p.State.Set(ctx, &m)...)
			return
		}
	}
	p.Diagnostics.AddError("Privilege Not Found", fmt.Sprintf("No privilege with ID %q was returned by OPNsense.", m.ID.ValueString()))
}
func stringList(s []string) types.List {
	v := make([]attr.Value, len(s))
	for i, x := range s {
		v[i] = types.StringValue(x)
	}
	return types.ListValueMust(types.StringType, v)
}

func (r *userPrivilegeResource) Metadata(ctx context.Context, q resource.MetadataRequest, p *resource.MetadataResponse) {
	p.TypeName = q.ProviderTypeName + "_auth_user_privilege"
}
func (r *userPrivilegeResource) Schema(ctx context.Context, q resource.SchemaRequest, p *resource.SchemaResponse) {
	p.Schema = privilegeAssignmentSchema("user")
}
func (r *userPrivilegeResource) Configure(ctx context.Context, q resource.ConfigureRequest, p *resource.ConfigureResponse) {
	if q.ProviderData != nil {
		configureData(&r.client, q.ProviderData, p.Diagnostics.AddError)
	}
}
func (r *groupPrivilegeResource) Metadata(ctx context.Context, q resource.MetadataRequest, p *resource.MetadataResponse) {
	p.TypeName = q.ProviderTypeName + "_auth_group_privilege"
}
func (r *groupPrivilegeResource) Schema(ctx context.Context, q resource.SchemaRequest, p *resource.SchemaResponse) {
	p.Schema = privilegeAssignmentSchema("group")
}
func (r *groupPrivilegeResource) Configure(ctx context.Context, q resource.ConfigureRequest, p *resource.ConfigureResponse) {
	if q.ProviderData != nil {
		configureData(&r.client, q.ProviderData, p.Diagnostics.AddError)
	}
}
func setAssignment(ctx context.Context, c opnsense.Client, m *privilegeAssignmentModel, group, add bool) error {
	x, e := c.Auth().PrivilegeGetItem(ctx, m.PrivilegeID.ValueString())
	if e != nil {
		return e
	}
	u := append([]string(nil), x.Item.Users...)
	g := append([]string(nil), x.Item.Groups...)
	if group {
		g = membership(g, m.PrincipalID.ValueString(), add)
	} else {
		u = membership(u, m.PrincipalID.ValueString(), add)
	}
	res, e := c.Auth().PrivilegeSetItem(ctx, m.PrivilegeID.ValueString(), &opnauth.PrivilegeSetItem{Users: strings.Join(u, ","), Groups: strings.Join(g, ",")})
	if e != nil {
		return e
	}
	if res.Result != "saved" {
		return fmt.Errorf("OPNsense returned %q", res.Result)
	}
	return nil
}
func membership(v []string, id string, add bool) []string {
	out := make([]string, 0, len(v)+1)
	found := false
	for _, x := range v {
		if x == id {
			found = true
			if !add {
				continue
			}
		}
		out = append(out, x)
	}
	if add && !found {
		out = append(out, id)
	}
	return out
}
func assignmentCreate(ctx context.Context, c opnsense.Client, group bool, q resource.CreateRequest, p *resource.CreateResponse) {
	var m privilegeAssignmentModel
	p.Diagnostics.Append(q.Plan.Get(ctx, &m)...)
	if p.Diagnostics.HasError() {
		return
	}
	if e := setAssignment(ctx, c, &m, group, true); e != nil {
		p.Diagnostics.AddError("Client Error", e.Error())
		return
	}
	m.ID = types.StringValue(m.PrivilegeID.ValueString() + ":" + m.PrincipalID.ValueString())
	p.Diagnostics.Append(p.State.Set(ctx, &m)...)
}
func assignmentRead(ctx context.Context, c opnsense.Client, group bool, q resource.ReadRequest, p *resource.ReadResponse) {
	var m privilegeAssignmentModel
	p.Diagnostics.Append(q.State.Get(ctx, &m)...)
	if p.Diagnostics.HasError() {
		return
	}
	x, e := c.Auth().PrivilegeGetItem(ctx, m.PrivilegeID.ValueString())
	if e != nil {
		p.Diagnostics.AddError("Client Error", e.Error())
		return
	}
	v := []string(x.Item.Users)
	if group {
		v = []string(x.Item.Groups)
	}
	for _, id := range v {
		if id == m.PrincipalID.ValueString() {
			p.Diagnostics.Append(p.State.Set(ctx, &m)...)
			return
		}
	}
	p.State.RemoveResource(ctx)
}
func assignmentDelete(ctx context.Context, c opnsense.Client, group bool, q resource.DeleteRequest, p *resource.DeleteResponse) {
	var m privilegeAssignmentModel
	p.Diagnostics.Append(q.State.Get(ctx, &m)...)
	if p.Diagnostics.HasError() {
		return
	}
	if e := setAssignment(ctx, c, &m, group, false); e != nil {
		p.Diagnostics.AddError("Client Error", e.Error())
	}
}
func (r *userPrivilegeResource) Create(ctx context.Context, q resource.CreateRequest, p *resource.CreateResponse) {
	assignmentCreate(ctx, r.client, false, q, p)
}
func (r *userPrivilegeResource) Read(ctx context.Context, q resource.ReadRequest, p *resource.ReadResponse) {
	assignmentRead(ctx, r.client, false, q, p)
}
func (r *userPrivilegeResource) Update(ctx context.Context, q resource.UpdateRequest, p *resource.UpdateResponse) {
}
func (r *userPrivilegeResource) Delete(ctx context.Context, q resource.DeleteRequest, p *resource.DeleteResponse) {
	assignmentDelete(ctx, r.client, false, q, p)
}
func (r *userPrivilegeResource) ImportState(ctx context.Context, q resource.ImportStateRequest, p *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), q, p)
}
func (r *groupPrivilegeResource) Create(ctx context.Context, q resource.CreateRequest, p *resource.CreateResponse) {
	assignmentCreate(ctx, r.client, true, q, p)
}
func (r *groupPrivilegeResource) Read(ctx context.Context, q resource.ReadRequest, p *resource.ReadResponse) {
	assignmentRead(ctx, r.client, true, q, p)
}
func (r *groupPrivilegeResource) Update(ctx context.Context, q resource.UpdateRequest, p *resource.UpdateResponse) {
}
func (r *groupPrivilegeResource) Delete(ctx context.Context, q resource.DeleteRequest, p *resource.DeleteResponse) {
	assignmentDelete(ctx, r.client, true, q, p)
}
func (r *groupPrivilegeResource) ImportState(ctx context.Context, q resource.ImportStateRequest, p *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), q, p)
}
