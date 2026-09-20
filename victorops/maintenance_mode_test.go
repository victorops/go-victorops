package victorops

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestGetMaintenanceModeState(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/maintenancemode", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{"companyId":"co-1","activeInstances":[{"instanceId":"mm-1","isGlobal":false,"startedAt":1620000000000,"startedBy":"alice","targets":[{"type":"RoutingKeys","names":["rk-web","rk-db"]}]}]}`))
	})

	state, _, err := testClient.GetMaintenanceModeState(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(state.ActiveInstances) != 1 || state.ActiveInstances[0].InstanceID != "mm-1" {
		t.Fatalf("unexpected state: %#v", state)
	}
	inst := state.ActiveInstances[0]
	if inst.StartedBy != "alice" || inst.StartedAt != 1620000000000 {
		t.Errorf("unexpected instance metadata: %#v", inst)
	}
	if len(inst.Targets) != 1 || inst.Targets[0].Type != "RoutingKeys" || len(inst.Targets[0].Names) != 2 {
		t.Errorf("expected nested targets to be parsed, got: %#v", inst.Targets)
	}
}

func TestStartMaintenanceModeWithKeys(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/maintenancemode/start", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		body, _ := io.ReadAll(r.Body)
		var payload StartMaintenanceModePayload
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("bad body: %v", err)
		}
		if payload.Type != "RoutingKeys" || len(payload.Names) != 1 || payload.Names[0] != "rk-web" {
			t.Errorf("unexpected payload: %#v", payload)
		}
		w.Write([]byte(`{"activeInstances":[{"instanceId":"mm-2","isGlobal":false,"targets":[{"type":"RoutingKeys","names":["rk-web"]}]}]}`))
	})

	state, _, err := testClient.StartMaintenanceMode(context.Background(), []string{"rk-web"}, "deploy")
	if err != nil {
		t.Fatal(err)
	}
	if len(state.ActiveInstances) != 1 || state.ActiveInstances[0].InstanceID != "mm-2" {
		t.Errorf("unexpected state: %#v", state)
	}
}

func TestStartMaintenanceModeGlobal(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/maintenancemode/start", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload StartMaintenanceModePayload
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("bad body: %v", err)
		}
		// Names must serialize as an empty array, never null, for global mode.
		if payload.Names == nil {
			t.Errorf("expected non-nil Names slice for global maintenance mode")
		}
		w.Write([]byte(`{"activeInstances":[{"instanceId":"mm-global","isGlobal":true}]}`))
	})

	state, _, err := testClient.StartMaintenanceMode(context.Background(), nil, "global")
	if err != nil {
		t.Fatal(err)
	}
	if state.ActiveInstances[0].InstanceID != "mm-global" {
		t.Errorf("unexpected state: %#v", state)
	}
}

func TestEndMaintenanceMode(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/maintenancemode/mm-2/end", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		w.Write([]byte(`{"activeInstances":[]}`))
	})

	state, _, err := testClient.EndMaintenanceMode(context.Background(), "mm-2")
	if err != nil {
		t.Fatal(err)
	}
	if len(state.ActiveInstances) != 0 {
		t.Errorf("expected no active instances, got %#v", state.ActiveInstances)
	}
}
