package cmd

import (
	"os"
	"testing"

	"github.com/google/uuid"
)

func TestGroupReadGuacamoleCmd(t *testing.T) {
	const endpoint = "http://127.0.0.1:54012"
	dbPath := uuid.New()
	e := testSetup(t, dbPath.String(), true)
	defer testCleanUp(dbPath.String())

	// Launch testing server
	go func() {
		e.Start(":54012")
	}()

	waitForTestServer(t, ":54012")

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
			name:           "guacamole group created",
			cmd:            NewGroupCmd(),
			args:           []string{serverFlag, endpoint, groupFlag, "test", descriptionFlag, "test", membersFlag, "kim,saul", "--guacamole-protocol", "ssh", guacParametersFlag, "host=192.168.1.1,port=22"},
			errorMessage:   "",
			successMessage: "Group created\n",
		},
		{
			name:           "list current groups",
			cmd:            ListGroupCmd(),
			args:           []string{serverFlag, endpoint, jsonFlag},
			errorMessage:   "",
			successMessage: `[{"gid":1,"name":"test","description":"test","members":[{"uid":3,"username":"saul","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","jpeg_photo":"","manager":false,"readonly":false,"locked":false},{"uid":4,"username":"kim","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","jpeg_photo":"","manager":false,"readonly":false,"locked":false}],"guac_config_protocol":"ssh","guac_config_parameters":"host=192.168.1.1,port=22"}]` + "\n",
		},
		{
			name:           "group test detail",
			cmd:            ListGroupCmd(),
			args:           []string{serverFlag, endpoint, "--gid", "1", jsonFlag},
			errorMessage:   "",
			successMessage: `{"gid":1,"name":"test","description":"test","members":[{"uid":3,"username":"saul","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","jpeg_photo":"","manager":false,"readonly":false,"locked":false},{"uid":4,"username":"kim","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","jpeg_photo":"","manager":false,"readonly":false,"locked":false}],"guac_config_protocol":"ssh","guac_config_parameters":"host=192.168.1.1,port=22"}` + "\n",
		},
	}

	for _, tc := range testCases {
		runTests(t, tc)
	}
}
