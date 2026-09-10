package victorops

import (
	"context"
	"net/http"
	"testing"
)

func TestCreateRoutingKey(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/org/routing-keys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.Write([]byte(`{
			"routingKey": "test-key",
			"targets": [ "pol-abcd" ]
		}`))
	})

	rk, _, err := testClient.CreateRoutingKey(context.Background(), &RoutingKey{
		RoutingKey: "test-key",
		Targets:    []string{"pol-abcd"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if rk.RoutingKey != "test-key" || len(rk.Targets) != 1 || rk.Targets[0] != "pol-abcd" {
		t.Errorf("unexpected routing key: %#v", rk)
	}
}

func TestGetAllRoutingKeys(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/org/routing-keys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{
			"routingKeys": [
				{ "routingKey": "test-key", "targets": [ { "policySlug": "pol-abcd" } ], "isDefault": true, "isMultiResponder": true },
				{ "routingKey": "other-key", "targets": [ { "policySlug": "pol-efgh" } ] }
			]
		}`))
	})

	list, _, err := testClient.GetAllRoutingKeys(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list.RoutingKeys) != 2 {
		t.Fatalf("expected 2 routing keys, got %d", len(list.RoutingKeys))
	}
	if list.RoutingKeys[0].Targets[0].PolicySlug != "pol-abcd" || !list.RoutingKeys[0].IsDefault || !list.RoutingKeys[0].IsMultiResponder {
		t.Errorf("unexpected first key targets: %#v", list.RoutingKeys[0])
	}
}

func TestGetRoutingKey(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/org/routing-keys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{
			"routingKeys": [
				{ "routingKey": "test-key", "targets": [ { "policySlug": "pol-abcd" } ] },
				{ "routingKey": "other-key", "targets": [ { "policySlug": "pol-efgh" } ] }
			]
		}`))
	})

	rk, _, err := testClient.GetRoutingKey(context.Background(), "other-key")
	if err != nil {
		t.Fatal(err)
	}
	if rk == nil || rk.RoutingKey != "other-key" || rk.Targets[0].PolicySlug != "pol-efgh" {
		t.Errorf("unexpected routing key: %#v", rk)
	}
}

func TestGetRoutingKeyNotFound(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/org/routing-keys", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{ "routingKeys": [] }`))
	})

	rk, _, err := testClient.GetRoutingKey(context.Background(), "missing-key")
	if err != nil {
		t.Fatal(err)
	}
	if rk != nil {
		t.Errorf("expected nil routing key for a missing name, got %#v", rk)
	}
}

func TestUpdateRoutingKey(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/org/routing-keys/test-key", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		w.Write([]byte(`{"routingKey":"test-key","targets":["pol-efgh"]}`))
	})

	rk, _, err := testClient.UpdateRoutingKey(context.Background(), "test-key", &RoutingKeyUpdatePayload{
		RoutingKey: "test-key",
		Targets:    []string{"pol-efgh"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if rk.RoutingKey != "test-key" || len(rk.Targets) != 1 || rk.Targets[0] != "pol-efgh" {
		t.Errorf("unexpected routing key: %#v", rk)
	}
}

func TestDeleteRoutingKey(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/org/routing-keys/test-key", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusOK)
	})

	details, err := testClient.DeleteRoutingKey(context.Background(), "test-key")
	if err != nil {
		t.Fatal(err)
	}
	if details.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", details.StatusCode)
	}
}
