package zenarmor

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func Resources(context.Context) []func() resource.Resource { return nil }

func DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{newStatusDataSource, newPoliciesDataSource, newInterfacesDataSource, newInterfaceDataSource}
}
