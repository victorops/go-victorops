package victorops

import (
	"context"
	"net/http"
	"testing"
)

func TestGetAlert(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/alerts/uuid-123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{
			"ackAuthor": "alice",
			"ackMsg": "on it",
			"entityDisplayName": "CPU high",
			"entityId": "host/cpu",
			"messageType": "CRITICAL",
			"monitoringTool": "nagios",
			"raw": "{\"foo\":\"bar\"}",
			"stateMessage": "cpu at 99%",
			"stateStartTime": 1620000000,
			"timestamp": 1620000123
		}`))
	})

	alert, _, err := testClient.GetAlert(context.Background(), "uuid-123")
	if err != nil {
		t.Fatal(err)
	}
	if alert.EntityID != "host/cpu" || alert.MessageType != "CRITICAL" {
		t.Errorf("unexpected alert: %#v", alert)
	}
	if alert.MonitoringTool != "nagios" || alert.AckAuthor != "alice" {
		t.Errorf("unexpected alert metadata: %#v", alert)
	}
	if alert.Raw != `{"foo":"bar"}` {
		t.Errorf("expected raw to be the original JSON string, got %q", alert.Raw)
	}
	if alert.Timestamp != 1620000123 || alert.StateStartTime != 1620000000 {
		t.Errorf("unexpected numeric timestamps: %#v", alert)
	}
}
