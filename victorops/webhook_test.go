package victorops

import (
	"context"
	"net/http"
	"testing"
)

func TestListWebhooks(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/webhooks", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{"webhooks":[{"label":"Slack","slug":"wh-1","url":"https://example.com/hook"}]}`))
	})

	list, _, err := testClient.ListWebhooks(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Webhooks) != 1 || list.Webhooks[0].Slug != "wh-1" || list.Webhooks[0].Label != "Slack" {
		t.Errorf("unexpected webhooks: %#v", list.Webhooks)
	}
}
