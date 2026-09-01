package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestResourcesIncludeIDS(t *testing.T) {
	t.Parallel()

	p := &opnsenseProvider{}
	found := make(map[string]bool)
	for _, factory := range p.Resources(context.Background()) {
		var metadata resource.MetadataResponse
		factory().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "opnsense"}, &metadata)
		found[metadata.TypeName] = true
	}

	for _, name := range []string{
		"opnsense_auth_user_api_key",
		"opnsense_ids_settings",
		"opnsense_ids_ruleset",
		"opnsense_ids_policy",
		"opnsense_ids_policy_rule",
		"opnsense_ids_user_rule",
		"opnsense_plugin",
	} {
		if !found[name] {
			t.Errorf("provider does not register %s", name)
		}
	}
}

func TestDataSourcesIncludeZenarmor(t *testing.T) {
	t.Parallel()
	p := &opnsenseProvider{}
	found := make(map[string]bool)
	for _, factory := range p.DataSources(context.Background()) {
		var metadata datasource.MetadataResponse
		factory().Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "opnsense"}, &metadata)
		found[metadata.TypeName] = true
	}
	for _, name := range []string{"opnsense_zenarmor_status", "opnsense_zenarmor_policies", "opnsense_zenarmor_interface", "opnsense_zenarmor_interfaces"} {
		if !found[name] {
			t.Errorf("provider does not register %s", name)
		}
	}
}
