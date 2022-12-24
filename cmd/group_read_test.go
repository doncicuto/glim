package cmd

import (
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/doncicuto/glim/common"
)

func TestGroupReadCmd(t *testing.T) {
	const endpoint = "http://127.0.0.1:51012"
	dbPath := uuid.New()
	e := testSetup(t, dbPath.String(), false)
	defer testCleanUp(dbPath.String())

	// Launch testing server
	go func() {
		e.Start(":51012")
	}()

	waitForTestServer(t, ":51012")

	// Get token path
	tokenPath, err := AuthTokenPath()
	if err != nil {
		t.Fatalf("could not get AuthTokenPath - %v", err)
	}
	os.Remove(tokenPath)

	testCases := []CmdTestCase{
		{
			name:           "login successful",
			cmd:            LoginCmd(),
			args:           []string{serverFlag, endpoint, usernameFlag, "admin", "--password", "test"},
			errorMessage:   "",
			successMessage: "Login succeeded\n",
		},
		{
			name:           "list initial groups",
			cmd:            ListGroupCmd(),
			args:           []string{serverFlag, endpoint, jsonFlag},
			errorMessage:   "",
			successMessage: `[]` + "\n",
		},
		{
			name:           "test group created",
			cmd:            NewGroupCmd(),
			args:           []string{serverFlag, endpoint, groupFlag, "test", descriptionFlag, "test", membersFlag, "saul,mike"},
			errorMessage:   "",
			successMessage: "Group created\n",
		},
		{
			name:           "list current groups",
			cmd:            ListGroupCmd(),
			args:           []string{serverFlag, endpoint, jsonFlag},
			errorMessage:   "",
			successMessage: `[{"gid":1,"name":"test","description":"test","members":[{"uid":3,"username":"saul","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","jpeg_photo":"","manager":false,"readonly":false,"locked":false},{"uid":5,"username":"mike","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","jpeg_photo":"","manager":false,"readonly":false,"locked":false}],"guac_config_protocol":"","guac_config_parameters":""}]` + "\n",
		},
		{
			name:           "group test detail",
			cmd:            ListGroupCmd(),
			args:           []string{serverFlag, endpoint, "--gid", "1", jsonFlag},
			errorMessage:   "",
			successMessage: `{"gid":1,"name":"test","description":"test","members":[{"uid":3,"username":"saul","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","jpeg_photo":"","manager":false,"readonly":false,"locked":false},{"uid":5,"username":"mike","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","jpeg_photo":"","manager":false,"readonly":false,"locked":false}],"guac_config_protocol":"","guac_config_parameters":""}` + "\n",
		},
		{
			name:           "login successful as kim",
			cmd:            LoginCmd(),
			args:           []string{serverFlag, endpoint, usernameFlag, "kim", "--password", "test"},
			errorMessage:   "",
			successMessage: "Login succeeded\n",
		},
		{
			name:           "kim can't get details about groups",
			cmd:            ListGroupCmd(),
			args:           []string{serverFlag, endpoint, jsonFlag},
			errorMessage:   common.UserHasNoProperPermissionsMessage,
			successMessage: "",
		},
		{
			name:           "kim can't get details about groups using ls",
			cmd:            ListGroupCmd(),
			args:           []string{serverFlag, endpoint, "ls", jsonFlag},
			errorMessage:   common.UserHasNoProperPermissionsMessage,
			successMessage: "",
		},
		{
			name:           "kim can't get details about specific group",
			cmd:            ListGroupCmd(),
			args:           []string{serverFlag, endpoint, "ls", "--gid", "3", jsonFlag},
			errorMessage:   common.UserHasNoProperPermissionsMessage,
			successMessage: "",
		},
	}

	for _, tc := range testCases {
		runTests(t, tc)
	}
}
