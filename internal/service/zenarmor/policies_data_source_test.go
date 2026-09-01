package zenarmor

import (
	"testing"

	clientzenarmor "github.com/browningluke/opnsense-go/pkg/zenarmor"
)

func TestPoliciesModelFromAPISortsByID(t *testing.T) {
	t.Parallel()
	model := policiesModelFromAPI([]clientzenarmor.Policy{{ID: 42, Name: "later"}, {ID: 0, Name: "Default", IsDefault: true}})
	if len(model.Policies) != 2 || model.Policies[0].ID.ValueInt64() != 0 || model.Policies[1].ID.ValueInt64() != 42 {
		t.Fatalf("unexpected policies: %#v", model.Policies)
	}
}
