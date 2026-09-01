package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/browningluke/opnsense-go/pkg/api"
	goauth "github.com/browningluke/opnsense-go/pkg/auth"
	"github.com/browningluke/opnsense-go/pkg/errs"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type userResource struct{ client opnsense.Client }
type groupResource struct{ client opnsense.Client }

func newUserResource() resource.Resource  { return &userResource{} }
func newGroupResource() resource.Resource { return &groupResource{} }
func configure(c *opnsense.Client, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	a, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *api.Client, got: %T", req.ProviderData))
		return
	}
	*c = opnsense.NewClient(a)
}
func (r *userResource) Metadata(ctx context.Context, q resource.MetadataRequest, p *resource.MetadataResponse) {
	p.TypeName = q.ProviderTypeName + "_auth_user"
}
func (r *userResource) Schema(ctx context.Context, q resource.SchemaRequest, p *resource.SchemaResponse) {
	p.Schema = userResourceSchema()
}
func (r *userResource) Configure(ctx context.Context, q resource.ConfigureRequest, p *resource.ConfigureResponse) {
	configure(&r.client, q, p)
}
func (r *userResource) Create(ctx context.Context, q resource.CreateRequest, p *resource.CreateResponse) {
	var d userResourceModel
	p.Diagnostics.Append(q.Plan.Get(ctx, &d)...)
	id, e := r.client.Auth().AddUser(ctx, userToAPI(d))
	if e != nil {
		p.Diagnostics.AddError("Client Error", e.Error())
		return
	}
	d.ID = types.StringValue(id)
	p.Diagnostics.Append(p.State.Set(ctx, &d)...)
}
func (r *userResource) Read(ctx context.Context, q resource.ReadRequest, p *resource.ReadResponse) {
	var d userResourceModel
	p.Diagnostics.Append(q.State.Get(ctx, &d)...)
	u, e := r.client.Auth().GetUser(ctx, d.ID.ValueString())
	if e != nil {
		var n *errs.NotFoundError
		if errors.As(e, &n) {
			p.State.RemoveResource(ctx)
			return
		}
		p.Diagnostics.AddError("Client Error", e.Error())
		return
	}
	d = userFromAPI(u, d)
	p.Diagnostics.Append(p.State.Set(ctx, &d)...)
}
func (r *userResource) Update(ctx context.Context, q resource.UpdateRequest, p *resource.UpdateResponse) {
	var d userResourceModel
	p.Diagnostics.Append(q.Plan.Get(ctx, &d)...)
	if e := r.client.Auth().UpdateUser(ctx, d.ID.ValueString(), userToAPI(d)); e != nil {
		p.Diagnostics.AddError("Client Error", e.Error())
		return
	}
	p.Diagnostics.Append(p.State.Set(ctx, &d)...)
}
func (r *userResource) Delete(ctx context.Context, q resource.DeleteRequest, p *resource.DeleteResponse) {
	var d userResourceModel
	p.Diagnostics.Append(q.State.Get(ctx, &d)...)
	if e := r.client.Auth().DeleteUser(ctx, d.ID.ValueString()); e != nil {
		p.Diagnostics.AddError("Client Error", e.Error())
	}
}
func (r *userResource) ImportState(ctx context.Context, q resource.ImportStateRequest, p *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), q, p)
}
func (r *groupResource) Metadata(ctx context.Context, q resource.MetadataRequest, p *resource.MetadataResponse) {
	p.TypeName = q.ProviderTypeName + "_auth_group"
}
func (r *groupResource) Schema(ctx context.Context, q resource.SchemaRequest, p *resource.SchemaResponse) {
	p.Schema = groupResourceSchema()
}
func (r *groupResource) Configure(ctx context.Context, q resource.ConfigureRequest, p *resource.ConfigureResponse) {
	configure(&r.client, q, p)
}
func (r *groupResource) Create(ctx context.Context, q resource.CreateRequest, p *resource.CreateResponse) {
	var d groupResourceModel
	p.Diagnostics.Append(q.Plan.Get(ctx, &d)...)
	id, e := r.client.Auth().AddGroup(ctx, groupToAPI(d))
	if e != nil {
		p.Diagnostics.AddError("Client Error", e.Error())
		return
	}
	d.ID = types.StringValue(id)
	p.Diagnostics.Append(p.State.Set(ctx, &d)...)
}
func (r *groupResource) Read(ctx context.Context, q resource.ReadRequest, p *resource.ReadResponse) {
	var d groupResourceModel
	p.Diagnostics.Append(q.State.Get(ctx, &d)...)
	g, e := r.client.Auth().GetGroup(ctx, d.ID.ValueString())
	if e != nil {
		var n *errs.NotFoundError
		if errors.As(e, &n) {
			p.State.RemoveResource(ctx)
			return
		}
		p.Diagnostics.AddError("Client Error", e.Error())
		return
	}
	d = groupFromAPI(g, d)
	p.Diagnostics.Append(p.State.Set(ctx, &d)...)
}
func (r *groupResource) Update(ctx context.Context, q resource.UpdateRequest, p *resource.UpdateResponse) {
	var d groupResourceModel
	p.Diagnostics.Append(q.Plan.Get(ctx, &d)...)
	if e := r.client.Auth().UpdateGroup(ctx, d.ID.ValueString(), groupToAPI(d)); e != nil {
		p.Diagnostics.AddError("Client Error", e.Error())
		return
	}
	p.Diagnostics.Append(p.State.Set(ctx, &d)...)
}
func (r *groupResource) Delete(ctx context.Context, q resource.DeleteRequest, p *resource.DeleteResponse) {
	var d groupResourceModel
	p.Diagnostics.Append(q.State.Get(ctx, &d)...)
	if e := r.client.Auth().DeleteGroup(ctx, d.ID.ValueString()); e != nil {
		p.Diagnostics.AddError("Client Error", e.Error())
	}
}
func (r *groupResource) ImportState(ctx context.Context, q resource.ImportStateRequest, p *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), q, p)
}

var _ = goauth.Group{}
