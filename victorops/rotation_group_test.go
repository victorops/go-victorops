package victorops

import (
	"context"
	"net/http"
	"testing"
)

func TestCreateRotationGroup(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/teams/team-a/rotations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.Write([]byte(`{"id":100,"label":"Primary","teamslug":"team-a"}`))
	})

	resp, _, err := testClient.CreateRotationGroup(context.Background(), "team-a", &RotationGroupCreatePayload{Label: "Primary"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ID != 100 || resp.Label != "Primary" {
		t.Errorf("unexpected group: %#v", resp)
	}
}

func TestGetRotationGroup(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/teams/team-a/rotations/100", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{"id":100,"label":"Primary"}`))
	})

	resp, _, err := testClient.GetRotationGroup(context.Background(), "team-a", 100)
	if err != nil {
		t.Fatal(err)
	}
	if resp.ID != 100 {
		t.Errorf("unexpected group: %#v", resp)
	}
}

func TestUpdateRotationGroup(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/teams/team-a/rotations/100", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		w.Write([]byte(`{"id":100,"label":"Renamed"}`))
	})

	resp, _, err := testClient.UpdateRotationGroup(context.Background(), "team-a", 100, &RotationGroupUpdatePayload{Label: "Renamed"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Label != "Renamed" {
		t.Errorf("unexpected group: %#v", resp)
	}
}

func TestDeleteRotationGroup(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/teams/team-a/rotations/100", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusOK)
	})

	details, err := testClient.DeleteRotationGroup(context.Background(), "team-a", 100)
	if err != nil {
		t.Fatal(err)
	}
	if details.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", details.StatusCode)
	}
}

func TestCreateRotationShift(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/teams/team-a/rotations/100", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.Write([]byte(`{"rot_id":200,"group_id":100,"label":"Week","shifttype":"std"}`))
	})

	resp, _, err := testClient.CreateRotationShift(context.Background(), "team-a", 100, &RotationShiftCreatePayload{
		Label:     "Week",
		Timezone:  "UTC",
		Start:     1000,
		Duration:  7,
		ShiftType: "std",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.RotID != 200 || resp.ShiftType != "std" {
		t.Errorf("unexpected shift: %#v", resp)
	}
}

func TestGetRotationShift(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/teams/team-a/rotations/100/200", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{"rot_id":200,"label":"Week","oncall":"johndoe"}`))
	})

	resp, _, err := testClient.GetRotationShift(context.Background(), "team-a", 100, 200)
	if err != nil {
		t.Fatal(err)
	}
	if resp.RotID != 200 || resp.OnCall != "johndoe" {
		t.Errorf("unexpected shift: %#v", resp)
	}
}

func TestUpdateRotationShift(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/teams/team-a/rotations/100/200", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		w.Write([]byte(`{"rot_id":200,"label":"Week2","duration":14}`))
	})

	resp, _, err := testClient.UpdateRotationShift(context.Background(), "team-a", 100, 200, &RotationShiftCreatePayload{
		Label:     "Week2",
		Timezone:  "UTC",
		Start:     1000,
		Duration:  14,
		ShiftType: "std",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Label != "Week2" || resp.Duration != 14 {
		t.Errorf("unexpected shift: %#v", resp)
	}
}

func TestDeleteRotationShift(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/teams/team-a/rotations/100/200", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusOK)
	})

	details, err := testClient.DeleteRotationShift(context.Background(), "team-a", 100, 200)
	if err != nil {
		t.Fatal(err)
	}
	if details.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", details.StatusCode)
	}
}

func TestAddRotationShiftMember(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/teams/team-a/rotations/100/200/members", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.WriteHeader(http.StatusOK)
	})

	details, err := testClient.AddRotationShiftMember(context.Background(), "team-a", 100, 200, &RotationMemberAddPayload{Username: "johndoe"})
	if err != nil {
		t.Fatal(err)
	}
	if details.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", details.StatusCode)
	}
}

func TestRemoveRotationShiftMember(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/teams/team-a/rotations/100/200/members", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusOK)
	})

	details, err := testClient.RemoveRotationShiftMember(context.Background(), "team-a", 100, 200, &RotationMemberRemovePayload{Username: "johndoe"})
	if err != nil {
		t.Fatal(err)
	}
	if details.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", details.StatusCode)
	}
}

func TestUpdateRotationShiftMemberPosition(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/teams/team-a/rotations/100/200/members", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		w.WriteHeader(http.StatusOK)
	})

	details, err := testClient.UpdateRotationShiftMemberPosition(context.Background(), "team-a", 100, 200, &RotationMemberPositionPayload{Username: "johndoe", Position: 2})
	if err != nil {
		t.Fatal(err)
	}
	if details.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", details.StatusCode)
	}
}

func TestGetScheduledShiftUser(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/teams/team-a/rotations/100/200/scheduled", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{"username":"johndoe","displayName":"John Doe","verified":true}`))
	})

	resp, _, err := testClient.GetScheduledShiftUser(context.Background(), "team-a", 100, 200)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Username != "johndoe" || !resp.Verified {
		t.Errorf("unexpected scheduled user: %#v", resp)
	}
}

func TestSetScheduledShiftUser(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/teams/team-a/rotations/100/200/scheduled", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		w.Write([]byte(`{"username":"janedoe","verified":true}`))
	})

	resp, _, err := testClient.SetScheduledShiftUser(context.Background(), "team-a", 100, 200, &SetScheduledShiftPayload{Username: "janedoe"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Username != "janedoe" {
		t.Errorf("unexpected scheduled user: %#v", resp)
	}
}
