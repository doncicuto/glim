package cmd

import (
	"testing"

	"github.com/google/uuid"

	"github.com/doncicuto/glim/common"
)

func TestNewGroupCmd(t *testing.T) {
	const endpoint = "http://127.0.0.1:51021"

	const descriptionFlag = descriptionFlag
	const membersFlag = membersFlag

	dbPath := uuid.New()
	e := testSetup(t, dbPath.String(), false)
	defer testCleanUp(dbPath.String())

	// Launch testing server
	go func() {
		e.Start(":51021")
	}()

	waitForTestServer(t, ":51021")

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
			args:           []string{serverFlag, endpoint, groupFlag, "test", descriptionFlag, "test", membersFlag, "kim,saul"},
			errorMessage:   "",
			successMessage: "Group created\n",
		},
		{
			name:           "group already exists",
			cmd:            NewGroupCmd(),
			args:           []string{serverFlag, endpoint, groupFlag, "test", descriptionFlag, "test", membersFlag, "kim,saul"},
			errorMessage:   "group already exists",
			successMessage: "",
		},
		{
			name:           "new group killers",
			cmd:            NewGroupCmd(),
			args:           []string{serverFlag, endpoint, groupFlag, "killers", descriptionFlag, "test", membersFlag, "charles"},
			errorMessage:   "",
			successMessage: "Group created\n",
		},
		{
			name:           "login successful as kim",
			cmd:            LoginCmd(),
			args:           []string{serverFlag, endpoint, usernameFlag, "kim", passwordFlag, "test"},
			errorMessage:   "",
			successMessage: "Login succeeded\n",
		},
		{
			name:           "kim can't add new group",
			cmd:            NewGroupCmd(),
			args:           []string{serverFlag, endpoint, groupFlag, "killers", descriptionFlag, "test", membersFlag, "charles"},
			errorMessage:   common.UserHasNoProperPermissionsMessage,
			successMessage: "",
		},
	}

	for _, tc := range testCases {
		runTests(t, tc)
	}
}
