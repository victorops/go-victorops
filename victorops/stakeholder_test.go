package victorops

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestSendStakeholderMessage(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/stakeholders/sendMessage", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		body, _ := io.ReadAll(r.Body)
		var sent StakeholderMessage
		if err := json.Unmarshal(body, &sent); err != nil {
			t.Fatalf("bad request body: %v", err)
		}
		if sent.Title != "Outage" || sent.Message != "investigating" || sent.IncidentID != 7 {
			t.Errorf("unexpected stakeholder payload: %#v", sent)
		}
		if len(sent.Recipients) != 1 || sent.Recipients[0] != "johndoe" {
			t.Errorf("unexpected stakeholder recipients: %#v", sent.Recipients)
		}
		w.WriteHeader(http.StatusOK)
	})

	details, err := testClient.SendStakeholderMessage(context.Background(), &StakeholderMessage{
		Title:          "Outage",
		Message:        "investigating",
		IncidentID:     7,
		Recipients:     []string{"johndoe"},
		RecipientTeams: []string{"team-a"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if details.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", details.StatusCode)
	}
}
