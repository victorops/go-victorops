package victorops

import "testing"

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
