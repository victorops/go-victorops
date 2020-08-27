# go-victorops
A VictorOps client for golang.

## Installation
`go get https://github.com/victorops/go-victorops.git`

## Important Note

This client is used to make the API calls that are mentioned here [VictorOps public API](https://help.victorops.com/knowledge-base/api/). However, some features (Rotations and Paging Policies) are not publicly available yet. 

## Example Usage
```go
package main

import (
	"fmt"

	"https://github.com/victorops/go-victorops.git"
)

func main() {

	// Client initialization
	victoropsClient := victorops.NewClient(apiID, apiKey, "https://api.victorops.com")

	// Get all users in an account
	userList, _, err := victoropsClient.GetAllUsers()
	if err != nil {
		panic(err)
	}

	// Create a new victorops team
	team := victorops.Team{
		Name: "Test Team",
	}
	newTeam, details, err := victoropsClient.CreateTeam(&team)
	if err != nil {
		panic(err)
	}

	if details.StatusCode != 200 {
		panic(fmt.Errorf("failed to create team (%d): %s", details.StatusCode, details.ResponseBody))
	}

	fmt.Printf("Created team: %s\n", newTeam.Name)
}
```
