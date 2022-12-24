package cmd

import (
	"testing"

	"github.com/google/uuid"
)

func TestUserDeleteCmd(t *testing.T) {
	const endpoint = "http://127.0.0.1:51008"

	dbPath := uuid.New()
	e := testSetup(t, dbPath.String(), false)
	defer testCleanUp(dbPath.String())

	// Launch testing server
	go func() {
		e.Start(":51008")
	}()

	waitForTestServer(t, ":51008")

	testCases := []CmdTestCase{
		{
			name:           "login successful",
			cmd:            LoginCmd(),
			args:           []string{serverFlag, endpoint, usernameFlag, "admin", passwordFlag, "test"},
			errorMessage:   "",
			successMessage: "Login succeeded\n",
		},
		{
			name:           "user created",
			cmd:            NewUserCmd(),
			args:           []string{serverFlag, endpoint, usernameFlag, "test", passwordFlag, "test"},
			errorMessage:   "",
			successMessage: "User created\n",
		},
		{
			name:           "user did not confirm deletion",
			cmd:            DeleteUserCmd(),
			args:           []string{serverFlag, endpoint, usernameFlag, "test"},
			errorMessage:   "ok, user wasn't deleted",
			successMessage: "",
		},
		{
			name:           "user deleted",
			cmd:            DeleteUserCmd(),
			args:           []string{serverFlag, endpoint, usernameFlag, "test", "--force"},
			errorMessage:   "",
			successMessage: "User account deleted\n",
		},
	}

	for _, tc := range testCases {
		runTests(t, tc)
	}
}
