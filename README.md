# go-victorops

A VictorOps (Splunk On-Call) client for Go.

This client wraps the [VictorOps public API](https://portal.victorops.com/api-docs/) and
provides a context-aware, resilient HTTP engine (client-side rate limiting, bounded
retries, and rich per-request diagnostics) on top of a broad set of resource operations.

## Requirements

- Go 1.22 or later

## Installation

```bash
go get "github.com/victorops/go-victorops/v2/victorops"
```

## Client engine

Every exported API operation takes a `context.Context` as its first argument and shares a
single hardened request engine:

- **Context-first** – all calls accept `context.Context` for cancellation and deadlines.
- **Client-side rate limiting** – a token-bucket limiter (`golang.org/x/time/rate`) defaults
  to 2 req/s, matching VictorOps API guidance.
- **Bounded retries with exponential backoff** – transient failures (`429, 500, 502, 503,
  504`) are retried for **idempotent** methods only (`GET/HEAD/PUT/DELETE/OPTIONS`).
  `POST`/`PATCH` are never auto-retried, to avoid duplicate side effects.
- **`Retry-After` awareness** – honors the server's `Retry-After` header (seconds or
  HTTP-date) on `429` responses.
- **Credential redaction** – `X-VO-Api-Id` / `X-VO-Api-Key` are masked in any request dumps.
- **Explicit API errors** – any non-2xx response is returned as an `*APIError` while still
  handing back the full `RequestDetails` for inspection.

## Quick start

```go
package main

import (
	"context"
	"fmt"

	"github.com/victorops/go-victorops/v2/victorops"
)

func main() {
	ctx := context.Background()

	// Client initialization
	client := victorops.NewClient(apiID, apiKey, "https://api.victorops.com")

	// Get all users in an account
	userList, _, err := client.GetAllUsers(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("found %d user page(s)\n", len(userList.Users))

	// Create a new victorops team
	team := victorops.Team{Name: "Test Team"}
	newTeam, _, err := client.CreateTeam(ctx, &team)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created team: %s\n", newTeam.Name)
}
```

## Configuration

Three constructors are available:

```go
// Defaults: 30s timeout, 2 req/s rate limit, 3 retries with exponential backoff.
client := victorops.NewClient(apiID, apiKey, "https://api.victorops.com")

// Supply your own http.Client (e.g. custom transport, proxy, timeouts).
client := victorops.NewConfigurableClient(apiID, apiKey, baseURL, http.Client{Timeout: 15 * time.Second})

// Full control over timeout, rate limit and retry behavior.
client := victorops.NewClientWithArgs(apiID, apiKey, baseURL, victorops.ClientArgs{
	TimeoutSeconds: 20,
	RateLimit:      2,   // requests per second
	RetryConfig:    nil, // use the default retry policy; MaxRetries: 0 explicitly disables retries
})
```

The active configuration can be inspected via `GetHTTPClient()`, `GetRateLimiter()`, and
`GetRetryConfig()`.

## Migrating from v1

Version 2 is a breaking upgrade. It requires Go 1.22 or later and uses the module path
`github.com/victorops/go-victorops/v2`. Update imports to
`github.com/victorops/go-victorops/v2/victorops` and add a `context.Context` as the first
argument to every API operation. Client construction remains compatible, while HTTP
failures are now returned as `*victorops.APIError` with a `*victorops.RequestDetails`
value for diagnostics. Calls are rate-limited by default, and eligible transient failures
may be retried according to the configured retry policy.

This is not an import-only migration. Some request and response models were aligned with
the public API: routing-key updates use `RoutingKeyUpdatePayload`, paging rules use a
nested `PagingRuleContact`, rotation writes use numeric IDs plus `RotationGroupMask` and
`MaskTime`, and maintenance-mode instances expose `Targets` rather than flattened routing
keys. Compile existing callers against v2 and adapt these signatures and models before
deploying the upgrade.

## Error handling & diagnostics

API operations expose a `*RequestDetails` return value alongside their typed result and
error:

```go
type RequestDetails struct {
	StatusCode    int
	ResponseBody  string
	RequestBody   string
	RawResponse   *http.Response
	RawRequest    *http.Request
	RetryCount    int           // retries performed for this call
	RateLimited   bool          // true if the server returned 429 at least once
	RetryAfter    time.Duration // server-provided Retry-After, when present
	ErrorCategory string        // "rate_limit" | "server_error" | "client_error" | "network"
}
```

Non-2xx responses are surfaced as an `*APIError` (which still returns `RequestDetails`):

```go
user, details, err := client.GetUser(ctx, "someuser")
if err != nil {
	var apiErr *victorops.APIError
	if errors.As(err, &apiErr) {
		// HTTP-level failure; inspect apiErr.StatusCode / details.ResponseBody
	}
	// otherwise a transport / decoding error
}
```

## API coverage

More than 100 operations across 18 resource areas.

### Users
- `CreateUser` – create a user
- `CreateUsersBatch` – create multiple users in one request
- `GetUser` – get a user by username
- `GetAllUsers` – list all users (v1)
- `GetAllUserV2` – list all users (v2)
- `GetUserByEmail` – look up a user by email
- `UpdateUser` – update a user
- `DeleteUser` – delete a user (with replacement reassignment)
- `GetUserDefaultEmailContactID` – get the user's default email contact ID

### Teams
- `CreateTeam` – create a team
- `GetTeam` – get a team by slug
- `GetAllTeams` – list all teams
- `UpdateTeam` – update a team (addressed by slug)
- `DeleteTeam` – delete a team
- `GetTeamMembers` – list team members
- `AddTeamMember` – add a member to a team
- `RemoveTeamMember` – remove a member (with replacement)
- `IsTeamMember` – check membership
- `GetTeamAdmins` – list team admins

### Contacts
- `CreateContact` – create a contact method
- `GetContact` – get a contact by external ID
- `GetAllContacts` – list all contacts for a user
- `GetContactByID` – get a contact by internal ID
- `UpdateDeviceContact` – update a user's contact device
- `DeleteContact` – delete a contact

### Escalation Policies
- `CreateEscalationPolicy` – create an escalation policy
- `UpdateEscalationPolicy` – replace a policy's mutable steps and paging-policy flag
- `GetEscalationPolicy` – get a policy by ID
- `GetAllEscalationPolicies` – list all policies
- `DeleteEscalationPolicy` – delete a policy

### Routing Keys
- `CreateRoutingKey` – create a routing key
- `GetRoutingKey` – get a routing key by name
- `GetAllRoutingKeys` – list all routing keys
- `UpdateRoutingKey` – update a routing key
- `DeleteRoutingKey` – delete a routing key

### On-Call
- `GetApiTeamSchedule` – get a team's on-call schedule (v2)
- `GetUserOnCallSchedule` – get a user's on-call schedule (v2)
- `TakeOnCallForTeam` – take on-call for a team
- `TakeOnCallForPolicy` – take on-call for a policy

### Incidents
- `GetIncident` – get an incident by ID
- `GetIncidents` – list incidents
- `CreateIncident` – create an incident
- `AcknowledgeIncidents` – acknowledge incidents
- `ResolveIncidents` – resolve incidents
- `AcknowledgeIncidentsByUser` – acknowledge all incidents for a user
- `ResolveIncidentsByUser` – resolve all incidents for a user
- `RerouteIncidents` – reroute incidents to different targets
- `GetIncidentNotes` – get notes for an incident
- `CreateIncidentNote` – add a note to an incident
- `UpdateIncidentNote` – update a note
- `DeleteIncidentNote` – delete a note

### Scheduled Overrides
- `ListScheduledOverrides` – list all scheduled overrides
- `CreateScheduledOverride` – create a scheduled override
- `GetScheduledOverride` – get an override by public ID
- `DeleteScheduledOverride` – delete an override
- `GetScheduledOverrideAssignments` – list assignments for an override
- `GetScheduledOverrideAssignment` – get a specific assignment
- `UpdateScheduledOverrideAssignment` – update an assignment
- `DeleteScheduledOverrideAssignment` – delete an assignment

### Maintenance Mode
- `GetMaintenanceModeState` – get current maintenance mode state
- `StartMaintenanceMode` – start maintenance mode (keyed or global)
- `EndMaintenanceMode` – end maintenance mode

### Alert Rules
- `ListAlertRules` – list global and routing-key-scoped alert rules
- `CreateAlertRule` – create an alert rule
- `GetAlertRule` – get an alert rule by ID
- `GetAlertRuleByUpdate` – compatibility fallback when a scoped rule is omitted from listing (uses PUT and is not side-effect-free)
- `UpdateAlertRule` – update an alert rule
- `DeleteAlertRule` – delete an alert rule

### Alerts
- `GetAlert` – get an alert by UUID

### Rotation listing (read-only)
- `ListRotationsV1` – list rotation groups for a team (v1)
- `GetRotationGroupByGroupID` – find a v1 rotation group by numeric ID
- `ListRotationsV2` – list rotations with details for a team (v2)
- `GetRotationByGroupID` – find a v2 rotation by numeric ID

### Rotation Groups
- `CreateRotationGroup` – create a rotation group (optionally with shifts)
- `CreateRotation` – create a group with optional paging controls and resolve its numeric ID and slug
- `GetRotationGroup` – get a rotation group by ID
- `UpdateRotationGroup` – rename a rotation group
- `DeleteRotationGroup` – delete a rotation group
- `DeleteRotation` – delete a rotation group using the established int64-ID API
- `CreateRotationShift` – add a shift to a group
- `GetRotationShift` – get a shift
- `UpdateRotationShift` – replace a shift
- `DeleteRotationShift` – delete a shift
- `AddRotationShiftMember` – add a member to a shift
- `RemoveRotationShiftMember` – remove a member from a shift
- `UpdateRotationShiftMemberPosition` – reposition a shift member
- `GetScheduledShiftUser` – get the scheduled user for a shift
- `SetScheduledShiftUser` – set the scheduled user for a shift

### Personal Paging Policies
- `GetUserPolicies` – get a user's paging policies summary
- `GetNotificationTypes` – list available notification types
- `GetPagingContactTypes` – list available contact types
- `GetTimeoutTypes` – list available timeout types
- `GetUserPagingPolicies` – get a user's paging policy steps (v1)
- `GetUserPagingPoliciesV2` – get a user's paging policies (v2)
- `CreatePagingPolicyStep` – create a step
- `CreatePagingPolicyStepWithPayload` – create a step with initial rules
- `GetPagingPolicyStep` – get a step
- `UpdatePagingPolicyStep` – update a step
- `UpdatePagingPolicyStepWithPayload` – replace a step using its complete timeout-and-rules payload
- `CreatePagingPolicyRule` – create a rule in a step
- `GetPagingPolicyRule` – get a rule
- `UpdatePagingPolicyRule` – update a rule
- `DeletePagingPolicyRule` – delete a rule

### Reporting
- `GetOnCallLog` – get the on-call log for a team
- `SearchIncidents` – search incident history
- `GetOnCallCurrent` – get current on-call across the org
- `GetUserTeams` – get the teams a user belongs to

### Chat
- `SendChatMessage` – post a chat message into VictorOps
- `GetChatMessages` – retrieve chat messages

### Stakeholders
- `SendStakeholderMessage` – send a message to stakeholders

### Webhooks
- `ListWebhooks` – list configured outgoing webhooks

## Authentication

All API calls require authentication via header:

- `X-VO-Api-Id`: your VictorOps API ID
- `X-VO-Api-Key`: your VictorOps API Key

These are set automatically by the client when you initialize it with a constructor.

## Testing

Unit tests use mocked HTTP servers and do not contact the live API:

```bash
go test -race ./...
```
