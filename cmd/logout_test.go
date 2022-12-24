package cmd

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestLogoutCmd(t *testing.T) {
	const endpoint = "http://127.0.0.1:51006"
	dbPath := uuid.New()
	e := testSetup(t, dbPath.String(), false)
	defer testCleanUp(dbPath.String())

	// Launch testing server
	go func() {
		e.Start(":51006")
	}()

	waitForTestServer(t, ":51006")

	t.Run("login successful", func(t *testing.T) {
		loginCmd := LoginCmd()
		b := bytes.NewBufferString("")
		loginCmd.SetOut(b)
		loginCmd.SetArgs([]string{serverFlag, endpoint, usernameFlag, "admin", passwordFlag, "test"})
		err := loginCmd.Execute()
		if err == nil {
			out, err := ioutil.ReadAll(b)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "Login succeeded\n", string(out))
		}
	})

	t.Run("logout successful", func(t *testing.T) {
		logoutCmd := LogoutCmd()
		b := bytes.NewBufferString("")
		logoutCmd.SetOut(b)
		logoutCmd.SetArgs([]string{serverFlag, endpoint})
		err := logoutCmd.Execute()
		if err == nil {
			out, err := ioutil.ReadAll(b)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "Removing login credentials\n", string(out))
		}
	})
}
