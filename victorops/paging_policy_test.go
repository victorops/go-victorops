package victorops

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestGetNotificationTypes(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/policies/types/notifications", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{"notificationTypes":[{"type":"email","description":"Email"}]}`))
	})

	resp, _, err := testClient.GetNotificationTypes(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.NotificationTypes) != 1 || resp.NotificationTypes[0].Type != "email" {
		t.Errorf("unexpected notification types: %#v", resp)
	}
}

func TestGetPagingContactTypes(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/policies/types/contacts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{"contactTypes":[{"type":"phone","description":"Phone"}]}`))
	})

	resp, _, err := testClient.GetPagingContactTypes(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.ContactTypes) != 1 || resp.ContactTypes[0].Type != "phone" {
		t.Errorf("unexpected contact types: %#v", resp)
	}
}

func TestGetTimeoutTypes(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/policies/types/timeouts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{"timeoutTypes":[{"type":"5","description":"5 minutes"}]}`))
	})

	resp, _, err := testClient.GetTimeoutTypes(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.TimeoutTypes) != 1 || resp.TimeoutTypes[0].Type != "5" {
		t.Errorf("unexpected timeout types: %#v", resp)
	}
}

func TestGetUserPagingPolicies(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/profile/johndoe/policies", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{"steps":[{"step":1,"timeout":5,"rules":[{"rule":1,"notificationType":"email"}]}]}`))
	})

	resp, _, err := testClient.GetUserPagingPolicies(context.Background(), "johndoe")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Steps) != 1 || resp.Steps[0].Timeout != 5 {
		t.Errorf("unexpected steps: %#v", resp)
	}
}

func TestGetUserPagingPoliciesV2(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v2/profile/johndoe/policies", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{"policies":[{"policyType":"primary","steps":[{"step":1,"timeout":5}]}]}`))
	})

	resp, _, err := testClient.GetUserPagingPoliciesV2(context.Background(), "johndoe")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Policies) != 1 || resp.Policies[0].PolicyType != "primary" {
		t.Errorf("unexpected policies: %#v", resp)
	}
}

func TestCreatePagingPolicyStep(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/profile/johndoe/policies", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.Write([]byte(`{"step":{"index":2,"timeout":10}}`))
	})

	resp, _, err := testClient.CreatePagingPolicyStep(context.Background(), "johndoe", 10)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Step.Index != 2 || resp.Step.Timeout != 10 {
		t.Errorf("unexpected step: %#v", resp)
	}
}

func TestGetPagingPolicyStep(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/profile/johndoe/policies/2", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{"step":{"index":2,"timeout":10,"rules":[{"index":0,"type":"email","contact":{"id":42,"type":"email"}}]}}`))
	})

	resp, _, err := testClient.GetPagingPolicyStep(context.Background(), "johndoe", 2)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Step.Index != 2 {
		t.Fatalf("unexpected step: %#v", resp)
	}
	if len(resp.Step.Rules) != 1 || resp.Step.Rules[0].Type != "email" || resp.Step.Rules[0].Contact.ID != 42 {
		t.Errorf("expected nested rule/contact to be parsed, got: %#v", resp.Step.Rules)
	}
}

func TestUpdatePagingPolicyStep(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/profile/johndoe/policies/2", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		w.Write([]byte(`{"step":{"index":2,"timeout":15}}`))
	})

	resp, _, err := testClient.UpdatePagingPolicyStep(context.Background(), "johndoe", 2, 15)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Step.Timeout != 15 {
		t.Errorf("unexpected step: %#v", resp)
	}
}

func TestCreatePagingPolicyRule(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/profile/johndoe/policies/2", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		// The request must send a nested contact ({id,type}) plus the
		// notification type, not a flat contactId/notificationType.
		body, _ := io.ReadAll(r.Body)
		var sent AddRulePayload
		if err := json.Unmarshal(body, &sent); err != nil {
			t.Fatalf("bad request body: %v", err)
		}
		if sent.Type != "email" || sent.Contact.ID != 42 || sent.Contact.Type != "email" {
			t.Errorf("unexpected request body: %#v", sent)
		}
		w.Write([]byte(`{"stepRule":{"index":1,"type":"email","contact":{"id":42,"type":"email"}}}`))
	})

	resp, _, err := testClient.CreatePagingPolicyRule(context.Background(), "johndoe", 2, "email", PagingRuleContact{ID: 42, Type: "email"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StepRule.Contact.ID != 42 || resp.StepRule.Type != "email" {
		t.Errorf("unexpected rule: %#v", resp)
	}
}

func TestGetPagingPolicyRule(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/profile/johndoe/policies/2/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{"stepRule":{"index":1,"type":"email","contact":{"id":42,"type":"email"}}}`))
	})

	resp, _, err := testClient.GetPagingPolicyRule(context.Background(), "johndoe", 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StepRule.Index != 1 || resp.StepRule.Contact.ID != 42 {
		t.Errorf("unexpected rule: %#v", resp)
	}
}

func TestUpdatePagingPolicyRule(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/profile/johndoe/policies/2/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		w.Write([]byte(`{"stepRule":{"index":1,"type":"sms","contact":{"id":7,"type":"phone"}}}`))
	})

	resp, _, err := testClient.UpdatePagingPolicyRule(context.Background(), "johndoe", 2, 1, "sms", PagingRuleContact{ID: 7, Type: "phone"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StepRule.Type != "sms" || resp.StepRule.Contact.ID != 7 {
		t.Errorf("unexpected rule: %#v", resp)
	}
}

func TestDeletePagingPolicyRule(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/profile/johndoe/policies/2/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.Write([]byte(`{"stepRule":{"index":1}}`))
	})

	resp, _, err := testClient.DeletePagingPolicyRule(context.Background(), "johndoe", 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StepRule.Index != 1 {
		t.Errorf("unexpected rule: %#v", resp)
	}
}

func TestGetUserPolicies(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/user/johndoe/policies", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{"username":"johndoe","userId":5,"policies":[{"order":1,"timeout":5,"contactType":"email","extId":"e-1"}]}`))
	})

	resp, _, err := testClient.GetUserPolicies(context.Background(), "johndoe")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Username != "johndoe" || len(resp.Policies) != 1 || resp.Policies[0].ContactType != "email" {
		t.Errorf("unexpected user policies: %#v", resp)
	}
}
