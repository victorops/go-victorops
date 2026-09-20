package victorops

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestCreateRotationGroup(t *testing.T) {
	setup()
	defer teardown()

	listCalls := 0
	testMux.HandleFunc("/api-public/v1/teams/team-a/rotations", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listCalls++
			if listCalls == 1 {
				w.Write([]byte(`{"rotationGroups":[]}`))
				return
			}
			w.Write([]byte(`{"rotationGroups":[{"teamSlug":"team-a","slug":"rtg-1","label":"Primary","groupId":100}]}`))
		case http.MethodPost:
			w.Write([]byte(`{"id":100,"label":"Primary","teamslug":"team-a"}`))
		default:
			t.Errorf("unexpected method %s", r.Method)
		}
	})

	resp, _, err := testClient.CreateRotationGroup(context.Background(), "team-a", &RotationGroupCreatePayload{Label: "Primary"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ID != 100 || resp.Label != "Primary" || resp.Slug != "rtg-1" {
		t.Errorf("unexpected group: %#v", resp)
	}
}

func TestCreateRotationGroupReturnsSlugResolutionErrorWhenCreateReturnsID(t *testing.T) {
	setup()
	defer teardown()

	listCalls := 0
	testMux.HandleFunc("/api-public/v1/teams/team-a/rotations", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listCalls++
			if listCalls == 1 {
				w.Write([]byte(`{"rotationGroups":[]}`))
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"lookup failed"}`))
		case http.MethodPost:
			w.Write([]byte(`{"id":100,"label":"Primary","teamslug":"team-a"}`))
		default:
			t.Errorf("unexpected method %s", r.Method)
		}
	})

	resp, _, err := testClient.CreateRotationGroup(context.Background(), "team-a", &RotationGroupCreatePayload{Label: "Primary"})
	if err == nil {
		t.Fatal("expected slug-resolution error")
	}
	if resp == nil || resp.ID != 100 || resp.Slug != "" {
		t.Errorf("unexpected partial group: %#v", resp)
	}
}

func TestCreateRotationGroupFallsBackToJodaDateAndResolvesID(t *testing.T) {
	setup()
	defer teardown()

	listCalls := 0
	postCalls := 0
	testMux.HandleFunc("/api-public/v1/teams/team-a/rotations", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listCalls++
			if listCalls == 1 {
				w.Write([]byte(`{"rotationGroups":[]}`))
				return
			}
			w.Write([]byte(`{"rotationGroups":[{"teamSlug":"team-a","slug":"rtg-2","label":"Fallback","groupId":101}]}`))
		case http.MethodPost:
			postCalls++
			var body struct {
				Shifts []map[string]interface{} `json:"shifts"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode create body: %v", err)
			}
			for _, field := range []string{"usernames", "shiftMembers"} {
				members, ok := body.Shifts[0][field].([]interface{})
				if !ok || len(members) != 1 || members[0] != "jane" {
					t.Errorf("request should normalize members into %s: %#v", field, body.Shifts[0])
				}
			}
			if postCalls == 1 {
				if _, ok := body.Shifts[0]["start"].(float64); !ok {
					t.Errorf("first request should use epoch milliseconds: %#v", body.Shifts[0]["start"])
				}
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error":"error.expected.jodadate.format"}`))
				return
			}
			if got, ok := body.Shifts[0]["start"].(string); !ok || got != "2026-09-10T00:00:00.000Z" {
				t.Errorf("fallback should use Joda-compatible timestamp, got %#v", body.Shifts[0]["start"])
			}
			w.Write([]byte(`{"label":"Fallback","shifts":[{"group_id":101,"rot_id":201,"shiftMembers":[{"slug":"mem-1","username":"jane"}]}]}`))
		default:
			t.Errorf("unexpected method %s", r.Method)
		}
	})

	start := int64(1788998400000) // 2026-09-10T00:00:00.000Z
	resp, _, err := testClient.CreateRotationGroup(context.Background(), "team-a", &RotationGroupCreatePayload{
		Label: "Fallback",
		Shifts: []RotationShiftCreatePayload{{
			Label:        "Primary",
			Timezone:     "UTC",
			Start:        start,
			Duration:     7,
			ShiftType:    "std",
			ShiftMembers: []string{"jane"},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if postCalls != 2 || resp.ID != 101 || resp.Slug != "rtg-2" || len(resp.Shifts) != 1 || len(resp.Shifts[0].ShiftMembers) != 1 {
		t.Errorf("unexpected fallback result: calls=%d response=%#v", postCalls, resp)
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
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode create shift body: %v", err)
		}
		for _, field := range []string{"usernames", "shiftMembers"} {
			members, ok := body[field].([]interface{})
			if !ok || len(members) != 1 || members[0] != "janedoe" {
				t.Errorf("create shift should normalize members into %s: %#v", field, body)
			}
		}
		w.Write([]byte(`{"rot_id":200,"group_id":100,"label":"Week","shifttype":"std"}`))
	})

	resp, _, err := testClient.CreateRotationShift(context.Background(), "team-a", 100, &RotationShiftCreatePayload{
		Label:        "Week",
		Timezone:     "UTC",
		Start:        1000,
		Duration:     7,
		ShiftType:    "std",
		ShiftMembers: []string{"janedoe"},
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
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode update shift body: %v", err)
		}
		for _, field := range []string{"usernames", "shiftMembers"} {
			members, ok := body[field].([]interface{})
			if !ok || len(members) != 1 || members[0] != "johndoe" {
				t.Errorf("update shift should preserve usernames in %s: %#v", field, body)
			}
		}
		w.Write([]byte(`{"rot_id":200,"label":"Week2","duration":14}`))
	})

	resp, _, err := testClient.UpdateRotationShift(context.Background(), "team-a", 100, 200, &RotationShiftCreatePayload{
		Label:     "Week2",
		Timezone:  "UTC",
		Start:     1000,
		Duration:  14,
		ShiftType: "std",
		Usernames: []string{"johndoe"},
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
