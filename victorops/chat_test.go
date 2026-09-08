package victorops

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestSendChatMessage(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/chat", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		body, _ := io.ReadAll(r.Body)
		var sent ChatMessage
		if err := json.Unmarshal(body, &sent); err != nil {
			t.Fatalf("bad request body: %v", err)
		}
		if sent.Text != "deploying" || sent.MonitoringTool != "ci" {
			t.Errorf("unexpected chat payload: %#v", sent)
		}
		w.WriteHeader(http.StatusOK)
	})

	details, err := testClient.SendChatMessage(context.Background(), &ChatMessage{
		Username:       "johndoe",
		Text:           "deploying",
		MonitoringTool: "ci",
	})
	if err != nil {
		t.Fatal(err)
	}
	if details.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", details.StatusCode)
	}
}

func TestGetChatMessages(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/chat", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		if got := r.URL.Query().Get("incidentId"); got != "42" {
			t.Errorf("expected incidentId=42, got %q", got)
		}
		if got := r.URL.Query().Get("limit"); got != "10" {
			t.Errorf("expected limit=10, got %q", got)
		}
		w.Write([]byte(`{"messages":[{"username":"johndoe","text":"hi","serviceTime":1000,"sequence":1}],"hasMore":true}`))
	})

	list, _, err := testClient.GetChatMessages(context.Background(), &GetChatMessagesOptions{IncidentID: 42, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Messages) != 1 || list.Messages[0].Username != "johndoe" || !list.HasMore {
		t.Errorf("unexpected chat list: %#v", list)
	}
}
