package cmd

import (
	"testing"

	"github.com/doncicuto/glim/common"
	"github.com/google/uuid"
)

func TestUpdateGuacamoleGroupCmd(t *testing.T) {
	const endpoint = "http://127.0.0.1:56022"

	dbPath := uuid.New()
	e := testSetup(t, dbPath.String(), true)
	defer testCleanUp(dbPath.String())

	// Launch testing server
	go func() {
		e.Start(":56022")
	}()

	waitForTestServer(t, ":56022")

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
			name:           "guacamole group updated",
			cmd:            UpdateGroupCmd(),
			args:           []string{serverFlag, endpoint, groupFlag, "test", descriptionFlag, "test", membersFlag, "kim,saul", "--guacamole-protocol", "vnc", guacParametersFlag, "host=192.168.1.1,port=22"},
			errorMessage:   "",
			successMessage: common.GroupUpdatedMessage,
		},
		{
			name:           "guacamole group can be updated again",
			cmd:            UpdateGroupCmd(),
			args:           []string{serverFlag, endpoint, groupFlag, "test", descriptionFlag, "test", membersFlag, "kim,saul", guacParametersFlag, "host=192.168.1.1,port=22"},
			errorMessage:   "",
			successMessage: common.GroupUpdatedMessage,
		},
		{
			name:           "new group test2",
			cmd:            NewGroupCmd(),
			args:           []string{serverFlag, endpoint, groupFlag, "test2", descriptionFlag, "test", membersFlag, "kim,saul"},
			errorMessage:   "",
			successMessage: "Group created\n",
		},
		{
			name:           "guacamole group can't be updated",
			cmd:            UpdateGroupCmd(),
			args:           []string{serverFlag, endpoint, groupFlag, "test2", descriptionFlag, "test", membersFlag, "kim,saul", guacParametersFlag, "host=192.168.1.1,port=22"},
			errorMessage:   "Apache Guacamole config protocol is required",
			successMessage: "",
		},
	}

	for _, tc := range testCases {
		runTests(t, tc)
	}
}
