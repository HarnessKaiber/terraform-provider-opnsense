package auth

import (
	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/auth"
	"github.com/browningluke/terraform-provider-opnsense/internal/tools"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type userResourceModel struct {
	Name           types.String `tfsdk:"name"`
	Password       types.String `tfsdk:"password"`
	Fullname       types.String `tfsdk:"fullname"`
	Email          types.String `tfsdk:"email"`
	Comment        types.String `tfsdk:"comment"`
	Expires        types.String `tfsdk:"expires"`
	AuthorizedKeys types.String `tfsdk:"authorized_keys"`
	ID             types.String `tfsdk:"id"`
	Disabled       types.Bool   `tfsdk:"disabled"`
}
type groupResourceModel struct {
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	SourceNetworks types.String `tfsdk:"source_networks"`
	ID             types.String `tfsdk:"id"`
}

func userResourceSchema() schema.Schema {
	return schema.Schema{Version: 1, MarkdownDescription: "Local OPNsense user account.", Attributes: map[string]schema.Attribute{
		"name":            schema.StringAttribute{MarkdownDescription: "Unique local username.", Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"password":        schema.StringAttribute{MarkdownDescription: "Optional local-login password. It is never returned by OPNsense.", Optional: true, Sensitive: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"disabled":        schema.BoolAttribute{MarkdownDescription: "When enabled, disables authentication for this user. Defaults to `false`.", Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
		"fullname":        schema.StringAttribute{MarkdownDescription: "Informational full name.", Optional: true},
		"email":           schema.StringAttribute{MarkdownDescription: "Informational email address.", Optional: true},
		"comment":         schema.StringAttribute{MarkdownDescription: "Informational comment.", Optional: true},
		"expires":         schema.StringAttribute{MarkdownDescription: "Optional account expiration date.", Optional: true},
		"authorized_keys": schema.StringAttribute{MarkdownDescription: "SSH authorized keys for this user.", Optional: true},
		"id":              schema.StringAttribute{MarkdownDescription: "UUID of the user.", Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}
func groupResourceSchema() schema.Schema {
	return schema.Schema{Version: 1, MarkdownDescription: "Local OPNsense group.", Attributes: map[string]schema.Attribute{
		"name":            schema.StringAttribute{MarkdownDescription: "Unique group name.", Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"description":     schema.StringAttribute{MarkdownDescription: "Informational group description.", Optional: true},
		"source_networks": schema.StringAttribute{MarkdownDescription: "Optional source-network restriction for group privileges.", Optional: true},
		"id":              schema.StringAttribute{MarkdownDescription: "UUID of the group.", Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}}
}
func userToAPI(d userResourceModel) *auth.User {
	return &auth.User{Name: d.Name.ValueString(), Password: d.Password.ValueString(), Disabled: tools.BoolToString(d.Disabled.ValueBool()), Fullname: d.Fullname.ValueString(), Email: d.Email.ValueString(), Comment: d.Comment.ValueString(), Expires: d.Expires.ValueString(), AuthorizedKeys: d.AuthorizedKeys.ValueString()}
}
func userFromAPI(u *auth.User, d userResourceModel) userResourceModel {
	d.Name = types.StringValue(u.Name)
	d.Disabled = types.BoolValue(tools.StringToBool(u.Disabled))
	d.Fullname = tools.StringOrNull(u.Fullname)
	d.Email = tools.StringOrNull(u.Email)
	d.Comment = tools.StringOrNull(u.Comment)
	d.Expires = tools.StringOrNull(u.Expires)
	d.AuthorizedKeys = tools.StringOrNull(u.AuthorizedKeys)
	return d
}
func groupToAPI(d groupResourceModel) *auth.Group {
	return &auth.Group{Name: d.Name.ValueString(), Description: d.Description.ValueString(), SourceNetworks: api.SelectedMap(d.SourceNetworks.ValueString())}
}
func groupFromAPI(g *auth.Group, d groupResourceModel) groupResourceModel {
	d.Name = types.StringValue(g.Name)
	d.Description = tools.StringOrNull(g.Description)
	d.SourceNetworks = tools.StringOrNull(g.SourceNetworks.String())
	return d
}
