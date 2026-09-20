package victorops

import (
	"context"
	"net/http"
	"testing"
)

func TestIncidents(t *testing.T) {
	tests := []struct {
		name       string
		JSONString string
	}{
		{
			name: "basic incident",
			JSONString: `
			{ "incidents": [ {
			"alertCount": 1,
			"currentPhase": "ACKED",
			"entityDisplayName": "Something to show",
			"entityId": "8ec50e06-4c90-4d4d-a2fd-2ffab1e69a63",
			"entityState": "CRITICAL",
			"entityType": "SERVICE",
			"incidentNumber": "4",
			"lastAlertId": "b522e157-867b-4c75-8361-66dcc6dc4479",
			"lastAlertTime": "2020-03-24T19:30:34Z",
			"pagedPolicies": [
				{
					"policy": {
						"name": "Example",
						"slug": "team-KXK4L1qPrbLWwa6w"
					},
					"team": {
						"name": "Example",
						"slug": "team-KXK4L1qPrbLWwa6w"
					}
				}
			],
			"pagedTeams": [
				"team-KXK4L1qPrbLWwa6w"
			],
			"pagedUsers": [],
			"routingKey": "routingdefault",
			"service": "Something to show",
			"startTime": "2020-03-24T19:30:34Z",
			"transitions": [
				{
					"at": "2020-03-24T19:31:57Z",
					"by": "taitken-stage",
					"name": "ACKED"
				}
			]
		} ] }`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			incidentList, err := parseIncidentsResponse(test.JSONString)
			if err != nil {
				t.Errorf("Parsing of incident failed with %s: %s", test.JSONString, err)
			}
			if len(incidentList.Incidents) < 1 {
				t.Errorf("Incidents list is empty: got: %v", incidentList)
			}

			var testIncident = incidentList.Incidents[0]
			if testIncident.AlertCount != 1 {
				t.Errorf("Incident alertCount is wrong: got: %v", testIncident)
			}

			if testIncident.EntityState != "CRITICAL" {
				t.Errorf("Incident EntityState is wrong: got: %v", testIncident.EntityState)
			}
		})
	}
}

func TestGetIncident(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/incidents/4", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{
			"alertCount": 1,
			"currentPhase": "ACKED",
			"entityState": "CRITICAL",
			"incidentNumber": "4",
			"transitions": [ { "name": "ACKED", "by": "janedoe", "at": "2020-03-24T19:31:57Z" } ]
		}`))
	})

	incident, _, err := testClient.GetIncident(context.Background(), 4)
	if err != nil {
		t.Fatal(err)
	}
	if incident.IncidentNumber != "4" || incident.EntityState != "CRITICAL" {
		t.Errorf("unexpected incident: %#v", incident)
	}
	if len(incident.Transitions) != 1 || incident.Transitions[0].Name != "ACKED" || incident.Transitions[0].By != "janedoe" {
		t.Errorf("unexpected transitions: %#v", incident.Transitions)
	}
}

func TestGetIncidents(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/incidents", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{
			"incidents": [
				{ "incidentNumber": "4", "entityState": "CRITICAL" },
				{ "incidentNumber": "5", "entityState": "WARNING" }
			]
		}`))
	})

	resp, _, err := testClient.GetIncidents(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Incidents) != 2 || resp.Incidents[1].IncidentNumber != "5" {
		t.Errorf("unexpected incidents: %#v", resp)
	}
}
