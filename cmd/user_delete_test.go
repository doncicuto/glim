package cmd

import (
	"testing"
)

func TestUserDeleteCmd(t *testing.T) {
	e := testSetup(t)
	defer testCleanUp()

	// Launch testing server
	go func() {
		e.Start(":51008")
	}()

	waitForTestServer(t, ":51008")

	testCases := []CmdTestCase{
		{
			name:           "login successful",
			cmd:            NewLoginCmd(),
			args:           []string{"--server", "http://127.0.0.1:51008", "--username", "admin", "--password", "test"},
			errorMessage:   "",
			successMessage: "Login succeeded\n",
		},
		{
			name:           "user created",
			cmd:            NewUserCmd(),
			args:           []string{"--server", "http://127.0.0.1:51008", "--username", "test", "--password", "test"},
			errorMessage:   "",
			successMessage: "User created\n",
		},
		{
			name:           "user did not confirm deletion",
			cmd:            DeleteUserCmd(),
			args:           []string{"--server", "http://127.0.0.1:51008", "--username", "test"},
			errorMessage:   "ok, user wasn't deleted",
			successMessage: "",
		},
		{
			name:           "user deleted",
			cmd:            DeleteUserCmd(),
			args:           []string{"--server", "http://127.0.0.1:51008", "--username", "test", "--force"},
			errorMessage:   "",
			successMessage: "User account deleted\n",
		},
	}

	for _, tc := range testCases {
		runTests(t, tc)
	}
}
