package zenarmor_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/browningluke/opnsense-go/pkg/api"
	clientcore "github.com/browningluke/opnsense-go/pkg/core"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
	"github.com/browningluke/terraform-provider-opnsense/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccZenarmorStatusDataSource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("acceptance test; set TF_ACC=1 to run")
	}
	acctest.AccPreCheck(t)

	packages, err := liveFirmwarePackages(t.Context())
	if err != nil {
		t.Fatalf("independent firmware API query failed: %s", err)
	}
	versions := installedZenarmorVersions(packages)
	if versions["os-sensei"] == "" {
		t.Fatal("os-sensei must be installed for the Zenarmor acceptance test")
	}
	status, err := liveClient().Zenarmor().Status(t.Context())
	if err != nil {
		if !isUnconfiguredStatusError(err) {
			t.Fatalf("independent Zenarmor status API query failed: %s", err)
		}
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acctest.AccPreCheck(t) },
			ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
			Steps: []resource.TestStep{{
				Config:      `data "opnsense_zenarmor_status" "test" {}`,
				ExpectError: regexp.MustCompile(`status code 460`),
			}},
		})
		return
	}
	featureCount := "0"
	if status.CloudThreatIntel {
		featureCount = "1"
	}
	databaseVersion := status.DatabaseVersion.Version
	if databaseVersion == "" {
		databaseVersion = status.Database.Info.Version
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `data "opnsense_zenarmor_status" "test" {}`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.opnsense_zenarmor_status.test", "id", "os-sensei"),
				resource.TestCheckResourceAttr("data.opnsense_zenarmor_status.test", "installed", "true"),
				resource.TestCheckResourceAttr("data.opnsense_zenarmor_status.test", "plugin_version", versions["os-sensei"]),
				resource.TestCheckResourceAttr("data.opnsense_zenarmor_status.test", "agent_version", versions["os-sensei-agent"]),
				resource.TestCheckResourceAttr("data.opnsense_zenarmor_status.test", "updater_version", versions["os-sensei-updater"]),
				resource.TestCheckResourceAttr("data.opnsense_zenarmor_status.test", "engine_version", status.Engine.Version),
				resource.TestCheckResourceAttr("data.opnsense_zenarmor_status.test", "engine_status", strconv.FormatBool(status.Eastpect.Status)),
				resource.TestCheckResourceAttr("data.opnsense_zenarmor_status.test", "application_database_version", databaseVersion),
				resource.TestCheckResourceAttr("data.opnsense_zenarmor_status.test", "edition", status.License),
				resource.TestCheckResourceAttr("data.opnsense_zenarmor_status.test", "license_status", status.License),
				resource.TestCheckResourceAttr("data.opnsense_zenarmor_status.test", "supported_features.#", featureCount),
				resource.TestCheckResourceAttr("data.opnsense_zenarmor_status.test", "tls_inspection_supported", "false"),
				resource.TestCheckResourceAttr("data.opnsense_zenarmor_status.test", "full_tls_inspection_supported", "false"),
				resource.TestCheckResourceAttr("data.opnsense_zenarmor_status.test", "cloud_access_supported", strconv.FormatBool(status.CloudThreatIntel)),
			),
		}},
	})
}

func TestAccZenarmorPoliciesDataSource(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("acceptance test; set TF_ACC=1 to run")
	}
	acctest.AccPreCheck(t)

	client := liveClient()
	policies, err := client.Zenarmor().Policies(t.Context())
	if err != nil {
		t.Fatalf("independent Zenarmor policy API query failed: %s", err)
	}
	sort.Slice(policies, func(i, j int) bool { return policies[i].ID < policies[j].ID })

	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("data.opnsense_zenarmor_policies.test", "id", "zenarmor-policies"),
		resource.TestCheckResourceAttr("data.opnsense_zenarmor_policies.test", "policies.#", strconv.Itoa(len(policies))),
	}
	for index, policy := range policies {
		prefix := fmt.Sprintf("policies.%d.", index)
		checks = append(checks,
			resource.TestCheckResourceAttr("data.opnsense_zenarmor_policies.test", prefix+"id", strconv.FormatInt(policy.ID, 10)),
			resource.TestCheckResourceAttr("data.opnsense_zenarmor_policies.test", prefix+"name", policy.Name),
			resource.TestCheckResourceAttr("data.opnsense_zenarmor_policies.test", prefix+"enabled", strconv.FormatBool(policy.IsActive)),
			resource.TestCheckResourceAttr("data.opnsense_zenarmor_policies.test", prefix+"default", strconv.FormatBool(policy.IsDefault)),
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `data "opnsense_zenarmor_policies" "test" {}`,
			Check:  resource.ComposeAggregateTestCheckFunc(checks...),
		}},
	})
}

func TestAccZenarmorInterfacesDataSources(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("acceptance test; set TF_ACC=1 to run")
	}
	acctest.AccPreCheck(t)
	status, err := liveClient().Zenarmor().Status(t.Context())
	if err != nil {
		if !isUnconfiguredStatusError(err) {
			t.Fatalf("independent Zenarmor status API query failed: %s", err)
		}
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acctest.AccPreCheck(t) },
			ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config:      `data "opnsense_zenarmor_interfaces" "test" {}`,
					ExpectError: regexp.MustCompile(`status code 460`),
				},
				{
					Config:      `data "opnsense_zenarmor_interface" "test" { name = "vtnet0" }`,
					ExpectError: regexp.MustCompile(`status code 460`),
				},
			},
		})
		return
	}
	if len(status.InterfacesList) == 0 {
		t.Fatal("Zenarmor must advertise at least one monitored interface for the acceptance test")
	}
	sort.Slice(status.InterfacesList, func(i, j int) bool { return status.InterfacesList[i].Interface < status.InterfacesList[j].Interface })
	selected := status.InterfacesList[0]
	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("data.opnsense_zenarmor_interfaces.test", "id", "zenarmor-interfaces"),
		resource.TestCheckResourceAttr("data.opnsense_zenarmor_interfaces.test", "interfaces.#", strconv.Itoa(len(status.InterfacesList))),
		resource.TestCheckResourceAttr("data.opnsense_zenarmor_interface.test", "id", selected.Interface),
		resource.TestCheckResourceAttr("data.opnsense_zenarmor_interface.test", "name", selected.Interface),
		resource.TestCheckResourceAttr("data.opnsense_zenarmor_interface.test", "enabled", strconv.FormatBool(status.AgentEnabled)),
		resource.TestCheckResourceAttr("data.opnsense_zenarmor_interface.test", "monitored", "true"),
		resource.TestCheckResourceAttr("data.opnsense_zenarmor_interface.test", "tags.#", strconv.Itoa(len(selected.Tags))),
	}
	for index, item := range status.InterfacesList {
		prefix := fmt.Sprintf("interfaces.%d.", index)
		checks = append(checks,
			resource.TestCheckResourceAttr("data.opnsense_zenarmor_interfaces.test", prefix+"id", item.Interface),
			resource.TestCheckResourceAttr("data.opnsense_zenarmor_interfaces.test", prefix+"name", item.Interface),
			resource.TestCheckResourceAttr("data.opnsense_zenarmor_interfaces.test", prefix+"monitored", "true"),
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`data "opnsense_zenarmor_interfaces" "test" {}
data "opnsense_zenarmor_interface" "test" { name = %q }`, selected.Interface),
			Check: resource.ComposeAggregateTestCheckFunc(checks...),
		}},
	})
}

func isUnconfiguredStatusError(err error) bool {
	return strings.Contains(err.Error(), "status code 460")
}

func liveFirmwarePackages(ctx context.Context) ([]clientcore.Package, error) {
	client := liveClient()
	info, err := client.Core().FirmwareInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("firmware info: %w", err)
	}
	return info.Plugin, nil
}

func liveClient() opnsense.Client {
	return opnsense.NewClient(api.NewClient(api.Options{
		Uri:           os.Getenv("OPNSENSE_URI"),
		APIKey:        os.Getenv("OPNSENSE_API_KEY"),
		APISecret:     os.Getenv("OPNSENSE_API_SECRET"),
		AllowInsecure: os.Getenv("OPNSENSE_ALLOW_INSECURE") == "true",
	}))
}

func installedZenarmorVersions(packages []clientcore.Package) map[string]string {
	versions := make(map[string]string)
	for _, pkg := range packages {
		if pkg.Installed == "1" {
			versions[pkg.Name] = pkg.Version
		}
	}
	return versions
}
