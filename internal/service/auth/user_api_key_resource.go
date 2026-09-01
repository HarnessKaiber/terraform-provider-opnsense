package auth

import (
	"context"
	"fmt"

	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &userAPIKeyResource{}
var _ resource.ResourceWithConfigure = &userAPIKeyResource{}
var _ resource.ResourceWithImportState = &userAPIKeyResource{}

func newUserAPIKeyResource() resource.Resource { return &userAPIKeyResource{} }

type userAPIKeyResource struct{ client opnsense.Client }

func (r *userAPIKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auth_user_api_key"
}

func (r *userAPIKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = userAPIKeyResourceSchema()
}

func (r *userAPIKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	apiClient, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}
	r.client = opnsense.NewClient(apiClient)
}

func (r *userAPIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data userAPIKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.Auth().UserAddApiKey(ctx, data.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create API key for user %q: %s", data.Username.ValueString(), err))
		return
	}
	if created.Result != "ok" || created.Key == "" || created.Secret == "" {
		resp.Diagnostics.AddError("API Key Creation Failed", fmt.Sprintf("OPNsense returned result %q while creating an API key for user %q.", created.Result, data.Username.ValueString()))
		return
	}

	keys, err := r.client.Auth().UserGetAllApiKeys(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("API key was created, but its identifier could not be retrieved: %s", err))
		return
	}
	for _, key := range keys.Rows {
		if key.Username == data.Username.ValueString() && key.Key == created.Key {
			data.ID = types.StringValue(key.Id)
			data.Key = types.StringValue(created.Key)
			data.Secret = types.StringValue(created.Secret)
			tflog.Trace(ctx, "created OPNsense API key", map[string]any{"username": data.Username.ValueString()})
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}
	resp.Diagnostics.AddError("API Key Creation Failed", "OPNsense created an API key but did not return it from the API-key list. The key may need to be revoked manually.")
}

func (r *userAPIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data userAPIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	keys, err := r.client.Auth().UserGetAllApiKeys(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list API keys: %s", err))
		return
	}
	for _, key := range keys.Rows {
		if key.Id == data.ID.ValueString() {
			data.Username = types.StringValue(key.Username)
			data.Key = types.StringValue(key.Key)
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}
	tflog.Warn(ctx, "OPNsense API key not present in remote, removing from state")
	resp.State.RemoveResource(ctx)
}

func (r *userAPIKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Unsupported API Key Update", "OPNsense API keys cannot be updated. Changing the owning username replaces the key.")
}

func (r *userAPIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data userAPIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := r.client.Auth().UserDeleteApiKey(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to revoke API key: %s", err))
		return
	}
	if result.Result != "deleted" && result.Result != "not found" {
		resp.Diagnostics.AddError("API Key Deletion Failed", fmt.Sprintf("OPNsense returned result %q while revoking the API key.", result.Result))
	}
}

func (r *userAPIKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
