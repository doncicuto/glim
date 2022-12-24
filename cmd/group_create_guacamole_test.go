package cmd

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewGuacamoleGroupCmd(t *testing.T) {
	dbPath := uuid.New()
	e := testSetup(t, dbPath.String(), true)
	defer testCleanUp(dbPath.String())

	// Launch testing server
	go func() {
		e.Start(":56021")
	}()

	waitForTestServer(t, ":56021")

	const endpoint = "http://127.0.0.1:56021"

	testCases := []CmdTestCase{
		{
			name:           "login successful",
			cmd:            LoginCmd(),
			args:           []string{serverFlag, endpoint, usernameFlag, "admin", passwordFlag, "test"},
			errorMessage:   "",
			successMessage: "Login succeeded\n",
		},
		{
			name:           "new group test",
			cmd:            NewGroupCmd(),
			args:           []string{serverFlag, endpoint, groupFlag, "test", descriptionFlag, "test", membersFlag, "kim,saul", "--guacamole-protocol", "ssh", guacParametersFlag, "host=192.168.1.1,port=22"},
			errorMessage:   "",
			successMessage: "Group created\n",
		},
		{
			name:           "guacamole protocol is required if parameters are set",
			cmd:            NewGroupCmd(),
			args:           []string{serverFlag, endpoint, groupFlag, "wrong", descriptionFlag, "test", membersFlag, "kim,saul", guacParametersFlag, "host=192.168.1.1,port=22"},
			errorMessage:   "Apache Guacamole config protocol is required",
			successMessage: "",
		},
	}

	for _, tc := range testCases {
		runTests(t, tc)
	}
}
