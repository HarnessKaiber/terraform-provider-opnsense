package auth

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{newUserResource, newGroupResource, newUserPrivilegeResource, newGroupPrivilegeResource, newUserAPIKeyResource}
}

func DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{newPrivilegeDataSource}
}
