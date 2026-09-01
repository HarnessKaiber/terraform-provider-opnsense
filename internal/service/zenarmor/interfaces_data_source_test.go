package zenarmor

import (
	"testing"

	clientzenarmor "github.com/browningluke/opnsense-go/pkg/zenarmor"
)

func TestInterfacesModelFromAPINormalizesOrdering(t *testing.T) {
	t.Parallel()
	model := interfacesModelFromAPI([]clientzenarmor.InterfaceInfo{
		{Interface: "vtnet2", Tags: []string{"z", "a"}},
		{Interface: "vtnet0"},
	}, true)
	if len(model.Interfaces) != 2 || model.Interfaces[0].Name.ValueString() != "vtnet0" || model.Interfaces[1].Name.ValueString() != "vtnet2" {
		t.Fatalf("unexpected interfaces: %#v", model.Interfaces)
	}
	if len(model.Interfaces[1].Tags) != 2 || model.Interfaces[1].Tags[0].ValueString() != "a" || !model.Interfaces[1].Enabled.ValueBool() || !model.Interfaces[1].Monitored.ValueBool() {
		t.Fatalf("unexpected interface normalization: %#v", model.Interfaces[1])
	}
}
