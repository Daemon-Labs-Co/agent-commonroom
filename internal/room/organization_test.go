package room

import (
	"reflect"
	"testing"
)

func TestOrganizationDecisionAuthority(t *testing.T) {
	t.Parallel()
	chart := OrganizationChart{Nodes: []OrganizationNode{{ActorID: "lead"}, {ActorID: "dev", ReportsTo: "lead"}, {ActorID: "junior", ReportsTo: "dev"}, {ActorID: "qa"}}}
	if !OrganizationallyOutranks(chart, "lead", "junior") || OrganizationallyOutranks(chart, "dev", "lead") || OrganizationallyOutranks(chart, "qa", "dev") || OrganizationallyOutranks(chart, "dev", "dev") {
		t.Fatal("incorrect reporting ancestry")
	}
	for _, tc := range []struct{ available, want []string }{{[]string{"lead", "dev", "junior", "qa"}, []string{"lead", "qa"}}, {[]string{"dev", "junior"}, []string{"dev"}}, {[]string{"junior", "qa", "qa"}, []string{"junior", "qa"}}} {
		if got := OrganizationDecisionPeers(chart, tc.available); !reflect.DeepEqual(got, tc.want) {
			t.Fatal(got, tc.want)
		}
	}
	cycle := OrganizationChart{Nodes: []OrganizationNode{{ActorID: "a", ReportsTo: "b"}, {ActorID: "b", ReportsTo: "a"}}}
	if OrganizationallyOutranks(cycle, "missing", "a") {
		t.Fatal("malformed chart must terminate")
	}
}
