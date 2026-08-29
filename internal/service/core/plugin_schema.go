package core

import (
	clientcore "github.com/browningluke/opnsense-go/pkg/core"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type pluginResourceModel struct {
	Name       types.String `tfsdk:"name"`
	Version    types.String `tfsdk:"version"`
	Repository types.String `tfsdk:"repository"`
	ID         types.String `tfsdk:"id"`
}

func pluginResourceSchema() schema.Schema {
	return schema.Schema{
		Version:             1,
		MarkdownDescription: "Installs and removes an OPNsense firmware plugin package. Package operations are asynchronous and Terraform waits for their inventory state to converge.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the firmware plugin package to install.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Installed plugin version.",
				Computed:            true,
			},
			"repository": schema.StringAttribute{
				MarkdownDescription: "Repository that provides the plugin.",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Plugin package name.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func modelFromPlugin(plugin *clientcore.Package) *pluginResourceModel {
	return &pluginResourceModel{
		Name:       types.StringValue(plugin.Name),
		Version:    types.StringValue(plugin.Version),
		Repository: types.StringValue(plugin.Repository),
		ID:         types.StringValue(plugin.Name),
	}
}
