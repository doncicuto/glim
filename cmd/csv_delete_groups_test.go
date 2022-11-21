package cmd

import (
	"os"
	"testing"

	"github.com/google/uuid"
)

func prepareGroupDeleteTestFiles() error {
	const csvGroupHeader = "gid, name\n"

	// File with new groups file
	f, err := os.Create(file0)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString("name,description,members,guac_config_protocol,guac_config_parameters\n")
	if err != nil {
		return err
	}
	_, err = f.WriteString(`"devel","Developers","saul,kim",,` + "\n")
	if err != nil {
		return err
	}
	_, err = f.WriteString(`"admins","Administratos","kim",,` + "\n")
	if err != nil {
		return err
	}
	f.Sync()

	// File with a good CSV file
	f, err = os.Create(file1)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(csvGroupHeader)
	if err != nil {
		return err
	}
	_, err = f.WriteString(`0,"devel"` + "\n")
	if err != nil {
		return err
	}
	_, err = f.WriteString(`2,""` + "\n")
	if err != nil {
		return err
	}
	f.Sync()

	// File with no groups
	f, err = os.Create(file2)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(csvGroupHeader)
	if err != nil {
		return err
	}
	f.Sync()

	// File with wrong header
	f, err = os.Create(file3)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString("gid, ndsdame\n")
	if err != nil {
		return err
	}
	_, err = f.WriteString(`"some","User","",` + "\n")
	if err != nil {
		return err
	}
	f.Sync()

	// File with groups with errors
	f, err = os.Create(file4)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(csvGroupHeader)
	if err != nil {
		return err
	}
	_, err = f.WriteString(`0,""` + "\n")
	if err != nil {
		return err
	}
	_, err = f.WriteString(`1,""` + "\n")
	if err != nil {
		return err
	}
	_, err = f.WriteString(`0,"devel"` + "\n")
	if err != nil {
		return err
	}
	f.Sync()

	return nil
}

func deleteGroupDeleteTestingFiles() {
	os.Remove(file0)
	os.Remove(file1)
	os.Remove(file2)
	os.Remove(file3)
	os.Remove(file4)
}

func TestCsvDeleteGroups(t *testing.T) {
	// Prepare test databases and echo testing server
	dbPath := uuid.New()
	e := testSetup(t, dbPath.String(), false)
	defer testCleanUp(dbPath.String())

	err := prepareGroupDeleteTestFiles()
	if err != nil {
		t.Fatal("error preparing CSV testing files")
	}
	defer deleteGroupDeleteTestingFiles()

	// Launch testing server
	go func() {
		e.Start(":50034")
	}()

	waitForTestServer(t, ":50034")
	const endpoint = "http://127.0.0.1:50034"
	const serverFlag = "--server"
	const fileFlag = "--file"

	testCases := []CmdTestCase{
		{
			name:           "login successful",
			cmd:            LoginCmd(),
			args:           []string{serverFlag, endpoint, "--username", "admin", "--password", "test"},
			errorMessage:   "",
			successMessage: "Login succeeded\n",
		},
		{
			name:           "file not found",
			cmd:            CsvDeleteGroupsCmd(),
			args:           []string{serverFlag, endpoint, fileFlag, "/tmp/file1"},
			errorMessage:   "can't open CSV file",
			successMessage: "",
		},
		{
			name:           "create groups should be succesful",
			cmd:            CsvCreateGroupsCmd(),
			args:           []string{serverFlag, endpoint, fileFlag, file0},
			errorMessage:   "",
			successMessage: "devel: successfully created\nadmins: successfully created\nCreate from CSV finished!\n",
		},
		{
			name:           "delete groups should be succesful",
			cmd:            CsvDeleteGroupsCmd(),
			args:           []string{serverFlag, endpoint, fileFlag, file1},
			errorMessage:   "",
			successMessage: "devel: successfully removed\n\n: successfully removed\n\nRemove from CSV finished!\n",
		},
		{
			name:           "group list should be empty",
			cmd:            ListGroupCmd(),
			args:           []string{serverFlag, endpoint, "--json"},
			errorMessage:   "",
			successMessage: `[]` + "\n",
		},
		{
			name:           "repeat file, groups should be skipped",
			cmd:            CsvDeleteGroupsCmd(),
			args:           []string{serverFlag, endpoint, fileFlag, file1},
			errorMessage:   "",
			successMessage: "devel: skipped, group not found\n\nGID 2: skipped, group not found\nRemove from CSV finished!\n",
		},
		{
			name:           "file without groups",
			cmd:            CsvDeleteGroupsCmd(),
			args:           []string{serverFlag, endpoint, fileFlag, file2},
			errorMessage:   "no groups where found in CSV file",
			successMessage: "",
		},
		{
			name:           "file with wrong header",
			cmd:            CsvDeleteGroupsCmd(),
			args:           []string{serverFlag, endpoint, fileFlag, file3},
			errorMessage:   "wrong header",
			successMessage: "",
		},
		{
			name:           "file with groups that have several errors",
			cmd:            CsvDeleteGroupsCmd(),
			args:           []string{serverFlag, endpoint, fileFlag, file4},
			errorMessage:   "",
			successMessage: "GID 0: skipped, invalid group name and gid\n\nGID 1: skipped, group not found\ndevel: skipped, group not found\n\nRemove from CSV finished!\n",
		},
	}

	for _, tc := range testCases {
		runTests(t, tc)
	}

}
