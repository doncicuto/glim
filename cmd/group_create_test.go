package cmd

import "testing"

func TestNewGroupCmd(t *testing.T) {
	e := testSetup(t)
	defer testCleanUp()

	// Launch testing server
	go func() {
		e.Start(":51011")
	}()

	waitForTestServer(t, ":51011")

	testCases := []CmdTestCase{
		{
			name:           "login successful",
			cmd:            LoginCmd(),
			args:           []string{"--server", "http://127.0.0.1:51011", "--username", "admin", "--password", "test"},
			errorMessage:   "",
			successMessage: "Login succeeded\n",
		},
		{
			name:           "new group test",
			cmd:            NewGroupCmd(),
			args:           []string{"--server", "http://127.0.0.1:51011", "--group", "test", "--description", "test", "--members", "kim,saul"},
			errorMessage:   "",
			successMessage: "Group created\n",
		},
		{
			name:           "group already exists",
			cmd:            NewGroupCmd(),
			args:           []string{"--server", "http://127.0.0.1:51011", "--group", "test", "--description", "test", "--members", "kim,saul"},
			errorMessage:   "group already exists",
			successMessage: "",
		},
		{
			name:           "new group killers",
			cmd:            NewGroupCmd(),
			args:           []string{"--server", "http://127.0.0.1:51011", "--group", "killers", "--description", "test", "--members", "charles"},
			errorMessage:   "",
			successMessage: "Group created\n",
		},
	}

	for _, tc := range testCases {
		runTests(t, tc)
	}
}
