package cmd

import (
	"testing"

	"github.com/google/uuid"

	"github.com/doncicuto/glim/common"
)

func TestUpdateGroupCmd(t *testing.T) {
	const endpoint = "http://127.0.0.1:51015"
	dbPath := uuid.New()
	e := testSetup(t, dbPath.String(), false)
	defer testCleanUp(dbPath.String())

	// Launch testing server
	go func() {
		e.Start(":51015")
	}()

	waitForTestServer(t, ":51015")

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
			name:           "update group requires gid or group name",
			cmd:            UpdateGroupCmd(),
			args:           []string{serverFlag, endpoint, descriptionFlag, "new description"},
			errorMessage:   "you must specify either the group id or name",
			successMessage: "",
		},
		{
			name:           "can't update group using non-existent group name",
			cmd:            UpdateGroupCmd(),
			args:           []string{serverFlag, endpoint, groupFlag, "what"},
			errorMessage:   "group not found",
			successMessage: "",
		},
		{
			name:           "can't update group using non-existent gid",
			cmd:            UpdateGroupCmd(),
			args:           []string{serverFlag, endpoint, "--gid", "10"},
			errorMessage:   "group not found",
			successMessage: "",
		},
		{
			name:           "update group description",
			cmd:            UpdateGroupCmd(),
			args:           []string{serverFlag, endpoint, groupFlag, "test", descriptionFlag, "new description"},
			errorMessage:   "",
			successMessage: common.GroupUpdatedMessage,
		},
		{
			name:           "group test detail is ok",
			cmd:            ListGroupCmd(),
			args:           []string{serverFlag, endpoint, "--gid", "1", jsonFlag},
			errorMessage:   "",
			successMessage: `{"gid":1,"name":"test","description":"new description","members":[{"uid":3,"username":"saul","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","jpeg_photo":"","manager":false,"readonly":false,"locked":false},{"uid":4,"username":"kim","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","jpeg_photo":"","manager":false,"readonly":false,"locked":false}],"guac_config_protocol":"","guac_config_parameters":""}` + "\n",
		},
		{
			name:           "add mike as group members",
			cmd:            UpdateGroupCmd(),
			args:           []string{serverFlag, endpoint, groupFlag, "test", membersFlag, "mike"},
			errorMessage:   "",
			successMessage: common.GroupUpdatedMessage,
		},
		{
			name:           "mike is a new member",
			cmd:            ListGroupCmd(),
			args:           []string{serverFlag, endpoint, "--gid", "1", jsonFlag},
			errorMessage:   "",
			successMessage: `{"gid":1,"name":"test","description":"new description","members":[{"uid":3,"username":"saul","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","jpeg_photo":"","manager":false,"readonly":false,"locked":false},{"uid":4,"username":"kim","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","jpeg_photo":"","manager":false,"readonly":false,"locked":false},{"uid":5,"username":"mike","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","jpeg_photo":"","manager":false,"readonly":false,"locked":false}],"guac_config_protocol":"","guac_config_parameters":""}` + "\n",
		},
		{
			name:           "replace all members",
			cmd:            UpdateGroupCmd(),
			args:           []string{serverFlag, endpoint, groupFlag, "test", membersFlag, "mike", "--replace"},
			errorMessage:   "",
			successMessage: common.GroupUpdatedMessage,
		},
		{
			name:           "mike is the only member",
			cmd:            ListGroupCmd(),
			args:           []string{serverFlag, endpoint, "--gid", "1", jsonFlag},
			errorMessage:   "",
			successMessage: `{"gid":1,"name":"test","description":"new description","members":[{"uid":5,"username":"mike","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","jpeg_photo":"","manager":false,"readonly":false,"locked":false}],"guac_config_protocol":"","guac_config_parameters":""}` + "\n",
		},
		{
			name:           "login successful as kim",
			cmd:            LoginCmd(),
			args:           []string{serverFlag, endpoint, usernameFlag, "kim", passwordFlag, "test"},
			errorMessage:   "",
			successMessage: "Login succeeded\n",
		},
		{
			name:           "kim can't update group",
			cmd:            UpdateGroupCmd(),
			args:           []string{serverFlag, endpoint, groupFlag, "test", descriptionFlag, "new description"},
			errorMessage:   common.UserHasNoProperPermissionsMessage,
			successMessage: "",
		},
	}

	for _, tc := range testCases {
		runTests(t, tc)
	}
}
