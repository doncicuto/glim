package cmd

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogoutCmd(t *testing.T) {
	e := testSetup(t)
	defer testCleanUp()

	// Launch testing server
	go func() {
		e.Start(":51006")
	}()

	waitForTestServer(t, ":51006")

	t.Run("login successful", func(t *testing.T) {
		loginCmd := NewLoginCmd()
		b := bytes.NewBufferString("")
		loginCmd.SetOut(b)
		loginCmd.SetArgs([]string{"--server", "http://127.0.0.1:51006", "--username", "admin", "--password", "test"})
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
		logoutCmd := NewLogoutCmd()
		b := bytes.NewBufferString("")
		logoutCmd.SetOut(b)
		logoutCmd.SetArgs([]string{"--server", "http://127.0.0.1:51006"})
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
