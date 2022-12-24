package cmd

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/doncicuto/glim/common"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestLoginCmd(t *testing.T) {
	const endpoint = "http://127.0.0.1:51005"
	dbPath := uuid.New()
	e := testSetup(t, dbPath.String(), false)
	defer testCleanUp(dbPath.String())

	// Launch testing server
	go func() {
		e.Start(":51005")
	}()

	waitForTestServer(t, ":51005")

	cmd := LoginCmd()

	t.Run("can't connect with server", func(t *testing.T) {
		cmd.SetArgs([]string{serverFlag, endpoint, usernameFlag, "admin", passwordFlag, "tess"})
		err := cmd.Execute()
		if err != nil {
			assert.Contains(t, err.Error(), "can't connect with Glim")
		}
	})

	t.Run("username is required", func(t *testing.T) {
		cmd.SetArgs([]string{serverFlag, endpoint, usernameFlag, ""})
		err := cmd.Execute()
		if err != nil {
			assert.Equal(t, "non-null username required", err.Error())
		}
	})

	t.Run(common.WrongUsernameOrPasswordMessage, func(t *testing.T) {
		cmd.SetArgs([]string{serverFlag, endpoint, usernameFlag, "admin", passwordFlag, "tess"})
		err := cmd.Execute()
		if err != nil {
			assert.Equal(t, common.WrongUsernameOrPasswordMessage, err.Error())
		}
	})

	t.Run("login successful", func(t *testing.T) {
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{serverFlag, endpoint, usernameFlag, "admin", passwordFlag, "test"})
		err := cmd.Execute()
		if err == nil {
			out, err := ioutil.ReadAll(b)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "Login succeeded\n", string(out))
		}
	})
}
