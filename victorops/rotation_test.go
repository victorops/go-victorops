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
		w.Write([]byte(`{"rotationGroups":[{"teamSlug":"team-a","slug":"rtg-1","label":"Primary","groupId":100}]}`))
	})

	list, _, err := testClient.ListRotationsV1(context.Background(), "team-a")
	if err != nil {
		t.Fatal(err)
	}
	if len(list.RotationGroups) != 1 || list.RotationGroups[0].Slug != "rtg-1" || list.RotationGroups[0].TeamSlug != "team-a" || list.RotationGroups[0].Label != "Primary" || list.RotationGroups[0].GroupID != 100 {
		t.Errorf("unexpected rotation groups: %#v", list.RotationGroups)
	}
}

func TestListRotationsV2(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v2/team/team-a/rotations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{"rotations":[{"groupId":100,"label":"Primary","totalMembersInRotation":2,"shifts":[{"shiftId":200,"label":"Week","duration":7,"start":"2026-09-10T00:00:00.000Z","timezone":"UTC","shifttype":"std","shiftMembers":[{"slug":"mem-1","username":"jane"}]}]}]}`))
	})

	list, _, err := testClient.ListRotationsV2(context.Background(), "team-a")
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Rotations) != 1 || list.Rotations[0].GroupID != 100 || list.Rotations[0].TotalMembersInRotation != 2 || len(list.Rotations[0].Shifts) != 1 || list.Rotations[0].Shifts[0].ShiftID != 200 || len(list.Rotations[0].Shifts[0].ShiftMembers) != 1 {
		t.Errorf("unexpected rotations: %#v", list.Rotations)
	}
}
