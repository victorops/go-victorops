package victorops

import (
	"context"
	"net/http"
	"testing"
)

func TestListRotationsV1(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/teams/team-a/rotations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{"rotationGroups":[{"name":"Primary","slug":"rtg-1"}]}`))
	})

	list, _, err := testClient.ListRotationsV1(context.Background(), "team-a")
	if err != nil {
		t.Fatal(err)
	}
	if len(list.RotationGroups) != 1 || list.RotationGroups[0].Slug != "rtg-1" {
		t.Errorf("unexpected rotation groups: %#v", list.RotationGroups)
	}
}

func TestListRotationsV2(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v2/team/team-a/rotations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{"rotations":[{"name":"Primary","slug":"rtg-1","shiftLength":7,"shiftLengthUnit":"days"}]}`))
	})

	list, _, err := testClient.ListRotationsV2(context.Background(), "team-a")
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Rotations) != 1 || list.Rotations[0].ShiftLength != 7 {
		t.Errorf("unexpected rotations: %#v", list.Rotations)
	}
}
