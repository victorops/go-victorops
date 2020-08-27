package victorops

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestGetApiTeamSchedule(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v2/team/teamSlug/oncall/schedule", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`
        {
          "team": {
            "name": "Infrastructure",
            "slug": "team-abcd"
          },
          "schedules": [
            {
              "policy": {
                "name": "High Severity",
                "slug": "pol-abcd"
              },
              "schedule": [
                {
                  "onCallUser": {
                    "username": "janedoe"
                  },
                  "onCallType": "rotation_group",
                  "rotationName": "Primary",
                  "shiftName": "primary",
                  "shiftRoll": "2020-03-31T09:00:00-06:00",
                  "rolls": [
                    {
                      "start": "2020-03-31T09:00:00-06:00",
                      "end": "2020-04-07T09:00:00-06:00",
                      "onCallUser": {
                        "username": "janedoe"
                      },
                      "isRoll": true
                    }
                  ]
                }
              ],
              "overrides": []
            }
          ]
        }
        `))
	})

	resp, _, err := testClient.GetApiTeamSchedule("teamSlug", 14, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	shiftStart, _ := time.Parse(time.RFC3339, "2020-03-31T09:00:00-06:00")
	shiftEnd, _ := time.Parse(time.RFC3339, "2020-04-07T09:00:00-06:00")
	want := &ApiTeamSchedule{
		Team: ApiTeam{
			Name: "Infrastructure",
			Slug: "team-abcd",
		},
		Schedules: []ApiEscalationPolicySchedule{
			{
				Policy: ApiEscalationPolicy{
					Name: "High Severity",
					Slug: "pol-abcd",
				},
				Schedule: []ApiOnCallEntry{
					{
						OnCallUser: ApiUser{
							Username: "janedoe",
						},
						OverrideOnCallUser: ApiUser{},
						OnCallType:         "rotation_group",
						RotationName:       "Primary",
						ShiftName:          "primary",
						ShiftRoll:          shiftStart,
						Rolls: []ApiOnCallRoll{
							{
								Start: shiftStart,
								End:   shiftEnd,
								OnCallUser: ApiUser{
									Username: "janedoe",
								},
								IsRoll: true,
							},
						},
					},
				},
				Overrides: []ApiOnCallOverride{},
			},
		},
	}

	if !reflect.DeepEqual(resp, want) {
		t.Errorf("returned \n\n%#v want \n\n%#v", resp, want)
	}
}

func TestGetUserSchedule(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v2/user/janedoe/oncall/schedule", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`
        {
          "teamSchedules": [
          {
              "team": {
                "name": "Infrastructure",
                "slug": "team-abcd"
              },
              "schedules": [
                {
                  "policy": {
                    "name": "High Severity",
                    "slug": "pol-abcd"
                  },
                  "schedule": [
                    {
                      "onCallUser": {
                        "username": "janedoe"
                      },
                      "onCallType": "rotation_group",
                      "rotationName": "Primary",
                      "shiftName": "primary",
                      "shiftRoll": "2020-03-31T09:00:00-06:00",
                      "rolls": [
                        {
                          "start": "2020-03-31T09:00:00-06:00",
                          "end": "2020-04-07T09:00:00-06:00",
                          "onCallUser": {
                            "username": "janedoe"
                          },
                          "isRoll": true
                        }
                      ]
                    }
                  ],
                  "overrides": []
                }
              ]
          }
          ]
        }
        `))
	})

	resp, _, err := testClient.GetUserOnCallSchedule("janedoe", 14, 0, 0)
	if err != nil {
		fmt.Println(err)
		t.Fatal(err)
	}

	shiftStart, _ := time.Parse(time.RFC3339, "2020-03-31T09:00:00-06:00")
	shiftEnd, _ := time.Parse(time.RFC3339, "2020-04-07T09:00:00-06:00")
	want := &ApiUserSchedule{
		Schedules: []ApiTeamSchedule{
			{
				Team: ApiTeam{
					Name: "Infrastructure",
					Slug: "team-abcd",
				},
				Schedules: []ApiEscalationPolicySchedule{
					{
						Policy: ApiEscalationPolicy{
							Name: "High Severity",
							Slug: "pol-abcd",
						},
						Schedule: []ApiOnCallEntry{
							{
								OnCallUser: ApiUser{
									Username: "janedoe",
								},
								OverrideOnCallUser: ApiUser{},
								OnCallType:         "rotation_group",
								RotationName:       "Primary",
								ShiftName:          "primary",
								ShiftRoll:          shiftStart,
								Rolls: []ApiOnCallRoll{
									{
										Start: shiftStart,
										End:   shiftEnd,
										OnCallUser: ApiUser{
											Username: "janedoe",
										},
										IsRoll: true,
									},
								},
							},
						},
						Overrides: []ApiOnCallOverride{},
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(resp, want) {
		t.Errorf("returned \n\n%#v want \n\n%#v", resp, want)
	}
}

func TestTakeOnCallForTeam(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/team/team-abcd/oncall/user", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PATCH")
		w.Write([]byte(`{"result":"ok"}`))
	})

	resp, _, err := testClient.TakeOnCallForTeam("team-abcd", &TakeRequest{
		FromUser: "janedoe",
		ToUser:   "johndoe",
	})
	if err != nil {
		fmt.Println(err)
		t.Fatal(err)
	}

	want := &TakeResponse{
		Result: "ok",
	}

	if !reflect.DeepEqual(resp, want) {
		t.Errorf("returned \n\n%#v want \n\n%#v", resp, want)
	}
}

func TestTakeOnCallForPolicy(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/policies/pol-abcd/oncall/user", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PATCH")
		w.Write([]byte(`{"result":"ok"}`))
	})

	resp, _, err := testClient.TakeOnCallForPolicy("pol-abcd", &TakeRequest{
		FromUser: "janedoe",
		ToUser:   "johndoe",
	})
	if err != nil {
		fmt.Println(err)
		t.Fatal(err)
	}

	want := &TakeResponse{
		Result: "ok",
	}

	if !reflect.DeepEqual(resp, want) {
		t.Errorf("returned \n\n%#v want \n\n%#v", resp, want)
	}
}
