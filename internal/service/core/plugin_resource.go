package core

import (
	"context"
	"fmt"
	"time"

	"github.com/browningluke/opnsense-go/pkg/api"
	clientcore "github.com/browningluke/opnsense-go/pkg/core"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

const pluginOperationTimeout = 10 * time.Minute

var _ resource.Resource = &pluginResource{}
var _ resource.ResourceWithConfigure = &pluginResource{}
var _ resource.ResourceWithImportState = &pluginResource{}

type pluginResource struct{ client opnsense.Client }

func newPluginResource() resource.Resource { return &pluginResource{} }
func (r *pluginResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_plugin"
}
func (r *pluginResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = pluginResourceSchema()
}
func (r *pluginResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *pluginResource) find(ctx context.Context, name string) (*clientcore.Package, error) {
	info, err := r.client.Core().FirmwareInfo(ctx)
	if err != nil {
		return nil, err
	}
	for _, plugin := range info.Plugin {
		if plugin.Name == name {
			return &plugin, nil
		}
	}
	return nil, nil
}
func (r *pluginResource) wait(ctx context.Context, name string, installed bool) (*clientcore.Package, error) {
	deadline := time.Now().Add(pluginOperationTimeout)
	for {
		plugin, err := r.find(ctx, name)
		if err != nil {
			return nil, err
		}
		isInstalled := plugin != nil && plugin.Installed == "1"
		if isInstalled == installed {
			return plugin, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timed out waiting for plugin %q installed=%t", name, installed)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
}
func (r *pluginResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var d pluginResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &d)...)
	if resp.Diagnostics.HasError() {
		return
	}
	p, err := r.find(ctx, d.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to query plugin inventory: %s", err))
		return
	}
	if p == nil || p.Provided != "1" {
		resp.Diagnostics.AddError("Plugin unavailable", fmt.Sprintf("Plugin %q is not available in the configured OPNsense repositories.", d.Name.ValueString()))
		return
	}
	if p.Installed != "1" {
		result, err := r.client.Core().FirmwareInstall(ctx, d.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to install plugin: %s", err))
			return
		}
		if result.Status != "ok" {
			resp.Diagnostics.AddError("Plugin install rejected", fmt.Sprintf("OPNsense returned status %q for plugin %q.", result.Status, d.Name.ValueString()))
			return
		}
	}
	p, err = r.wait(ctx, d.Name.ValueString(), true)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to confirm plugin installation: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, modelFromPlugin(p))...)
}
func (r *pluginResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var d pluginResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &d)...)
	if resp.Diagnostics.HasError() {
		return
	}
	p, err := r.find(ctx, d.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to query plugin inventory: %s", err))
		return
	}
	if p == nil || p.Installed != "1" {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, modelFromPlugin(p))...)
}
func (r *pluginResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {}
func (r *pluginResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var d pluginResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &d)...)
	if resp.Diagnostics.HasError() {
		return
	}
	p, err := r.find(ctx, d.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to query plugin inventory: %s", err))
		return
	}
	if p == nil || p.Installed != "1" {
		return
	}
	result, err := r.client.Core().FirmwareRemove(ctx, d.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove plugin: %s", err))
		return
	}
	if result.Status != "ok" {
		resp.Diagnostics.AddError("Plugin removal rejected", fmt.Sprintf("OPNsense returned status %q for plugin %q.", result.Status, d.Name.ValueString()))
		return
	}
	if _, err = r.wait(ctx, d.Name.ValueString(), false); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to confirm plugin removal: %s", err))
	}
}
func (r *pluginResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
