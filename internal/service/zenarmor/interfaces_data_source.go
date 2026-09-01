package zenarmor

import (
	"context"
	"fmt"
	"sort"

	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
	clientzenarmor "github.com/browningluke/opnsense-go/pkg/zenarmor"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type interfaceModel struct {
	ID          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	Description types.String   `tfsdk:"description"`
	Enabled     types.Bool     `tfsdk:"enabled"`
	Monitored   types.Bool     `tfsdk:"monitored"`
	Tags        []types.String `tfsdk:"tags"`
}

type interfacesModel struct {
	ID         types.String     `tfsdk:"id"`
	Interfaces []interfaceModel `tfsdk:"interfaces"`
}

type interfacesDataSource struct{ client opnsense.Client }
type interfaceDataSource struct{ client opnsense.Client }

func newInterfacesDataSource() datasource.DataSource { return &interfacesDataSource{} }
func newInterfaceDataSource() datasource.DataSource  { return &interfaceDataSource{} }

func (d *interfaceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zenarmor_interface"
}

func (d *interfaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := interfaceAttributes()
	attributes["name"] = schema.StringAttribute{Required: true, MarkdownDescription: "Exact OPNsense interface name to resolve."}
	resp.Schema = schema.Schema{MarkdownDescription: "Resolves one interface exposed by the local Zenarmor status controller.", Attributes: attributes}
}

func (d *interfaceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	apiClient, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *api.Client, got %T.", req.ProviderData))
		return
	}
	d.client = opnsense.NewClient(apiClient)
}

func (d *interfaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config interfaceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	status, err := d.client.Zenarmor().Status(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to query Zenarmor interfaces", err.Error())
		return
	}
	model := interfacesModelFromAPI(status.InterfacesList, status.AgentEnabled)
	for _, candidate := range model.Interfaces {
		if candidate.Name.ValueString() == config.Name.ValueString() {
			resp.Diagnostics.Append(resp.State.Set(ctx, &candidate)...)
			return
		}
	}
	resp.Diagnostics.AddError("Zenarmor interface not found", fmt.Sprintf("No Zenarmor-monitored interface named %q was returned by the local controller.", config.Name.ValueString()))
}

func (d *interfacesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zenarmor_interfaces"
}

func interfaceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Stable local interface identifier."},
		"name":        schema.StringAttribute{Computed: true, MarkdownDescription: "OPNsense interface name."},
		"description": schema.StringAttribute{Computed: true, MarkdownDescription: "Interface description when advertised by Zenarmor."},
		"enabled":     schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the Zenarmor agent is enabled."},
		"monitored":   schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the interface appears in Zenarmor's monitored interface list."},
		"tags":        schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Deterministically ordered Zenarmor tags."},
	}
}

func (d *interfacesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Lists interfaces exposed by the local Zenarmor status controller.", Attributes: map[string]schema.Attribute{
		"id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this interface collection."},
		"interfaces": schema.ListNestedAttribute{Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: interfaceAttributes()}},
	}}
}

func (d *interfacesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	apiClient, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *api.Client, got %T.", req.ProviderData))
		return
	}
	d.client = opnsense.NewClient(apiClient)
}

func (d *interfacesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	status, err := d.client.Zenarmor().Status(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Zenarmor interfaces", err.Error())
		return
	}
	model := interfacesModelFromAPI(status.InterfacesList, status.AgentEnabled)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func interfacesModelFromAPI(items []clientzenarmor.InterfaceInfo, enabled bool) interfacesModel {
	sort.Slice(items, func(i, j int) bool { return items[i].Interface < items[j].Interface })
	model := interfacesModel{ID: types.StringValue("zenarmor-interfaces"), Interfaces: make([]interfaceModel, 0, len(items))}
	for _, item := range items {
		sort.Strings(item.Tags)
		tags := make([]types.String, 0, len(item.Tags))
		for _, tag := range item.Tags {
			tags = append(tags, types.StringValue(tag))
		}
		model.Interfaces = append(model.Interfaces, interfaceModel{ID: types.StringValue(item.Interface), Name: types.StringValue(item.Interface), Description: types.StringNull(), Enabled: types.BoolValue(enabled), Monitored: types.BoolValue(true), Tags: tags})
	}
	return model
}
