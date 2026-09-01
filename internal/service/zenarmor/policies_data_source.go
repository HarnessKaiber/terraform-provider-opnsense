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

var _ datasource.DataSource = &policiesDataSource{}
var _ datasource.DataSourceWithConfigure = &policiesDataSource{}

type policyDataSourceModel struct {
	ID            types.Int64  `tfsdk:"id"`
	LocalID       types.Int64  `tfsdk:"local_id"`
	Name          types.String `tfsdk:"name"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Default       types.Bool   `tfsdk:"default"`
	Centralized   types.Bool   `tfsdk:"centralized"`
	CloudPolicyID types.String `tfsdk:"cloud_policy_id"`
}

type policiesDataSourceModel struct {
	ID       types.String            `tfsdk:"id"`
	Policies []policyDataSourceModel `tfsdk:"policies"`
}

type policiesDataSource struct{ client opnsense.Client }

func newPoliciesDataSource() datasource.DataSource { return &policiesDataSource{} }

func (d *policiesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zenarmor_policies"
}

func (d *policiesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists policies exposed by the local OPNsense Zenarmor controller.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this policy collection."},
			"policies": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"id":              schema.Int64Attribute{Computed: true, MarkdownDescription: "Zenarmor policy identifier."},
					"local_id":        schema.Int64Attribute{Computed: true, MarkdownDescription: "Local Zenarmor policy identifier."},
					"name":            schema.StringAttribute{Computed: true, MarkdownDescription: "Policy name."},
					"enabled":         schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the policy is active."},
					"default":         schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this is Zenarmor's default policy."},
					"centralized":     schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether Zenconsole centrally manages the policy."},
					"cloud_policy_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Zenconsole policy identifier, or an empty string for a local policy."},
				}},
			},
		},
	}
}

func (d *policiesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *policiesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	policies, err := d.client.Zenarmor().Policies(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Zenarmor policies", err.Error())
		return
	}
	model := policiesModelFromAPI(policies)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func policiesModelFromAPI(policies []clientzenarmor.Policy) policiesDataSourceModel {
	sort.Slice(policies, func(i, j int) bool { return policies[i].ID < policies[j].ID })
	model := policiesDataSourceModel{ID: types.StringValue("zenarmor-policies"), Policies: make([]policyDataSourceModel, 0, len(policies))}
	for _, policy := range policies {
		model.Policies = append(model.Policies, policyDataSourceModel{
			ID: types.Int64Value(policy.ID), LocalID: types.Int64Value(policy.LocalID), Name: types.StringValue(policy.Name),
			Enabled: types.BoolValue(policy.IsActive), Default: types.BoolValue(policy.IsDefault), Centralized: types.BoolValue(policy.IsCentralized),
			CloudPolicyID: types.StringValue(policy.CloudPolicyID),
		})
	}
	return model
}
