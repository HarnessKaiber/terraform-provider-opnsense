package zenarmor

import (
	"testing"

	clientcore "github.com/browningluke/opnsense-go/pkg/core"
	clientzenarmor "github.com/browningluke/opnsense-go/pkg/zenarmor"
)

func TestStatusModelFromPackages(t *testing.T) {
	t.Parallel()
	model := statusModelFromPackages([]clientcore.Package{
		{Name: "os-sensei", Version: "2.6.2", Installed: "1"},
		{Name: "os-sensei-agent", Version: "2.6.1", Installed: "1"},
		{Name: "os-sensei-updater", Version: "2.0", Installed: "1"},
		{Name: "os-unrelated", Version: "1.0", Installed: "1"},
	})
	if !model.Installed.ValueBool() || model.PluginVersion.ValueString() != "2.6.2" || model.AgentVersion.ValueString() != "2.6.1" || model.UpdaterVersion.ValueString() != "2.0" {
		t.Fatalf("unexpected model: %#v", model)
	}
}

func TestStatusModelFromAPI(t *testing.T) {
	t.Parallel()
	model := statusModelFromAPI(statusModelFromPackages([]clientcore.Package{{Name: "os-sensei", Version: "2.6.2", Installed: "1"}}), &clientzenarmor.Status{
		Engine: clientzenarmor.VersionInfo{Version: "1.18"}, Eastpect: clientzenarmor.ServiceStatus{Status: true},
		DatabaseVersion: clientzenarmor.VersionInfo{Version: "2026.08"}, License: "Free", CloudThreatIntel: true,
	})
	if model.EngineVersion.ValueString() != "1.18" || !model.EngineStatus.ValueBool() || model.ApplicationDatabaseVersion.ValueString() != "2026.08" {
		t.Fatalf("unexpected model: %#v", model)
	}
	if model.Edition.ValueString() != "Free" || len(model.SupportedFeatures) != 1 || model.SupportedFeatures[0].ValueString() != "cloud_threat_intelligence" {
		t.Fatalf("unexpected capabilities: %#v", model)
	}
}

func TestStatusModelRequiresInstalledPackage(t *testing.T) {
	t.Parallel()
	model := statusModelFromPackages([]clientcore.Package{{Name: "os-sensei", Version: "2.6.2", Installed: "0"}})
	if model.Installed.ValueBool() || !model.PluginVersion.IsNull() {
		t.Fatalf("unexpected model: %#v", model)
	}
}
