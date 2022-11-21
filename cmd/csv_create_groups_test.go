package cmd

import (
	"os"
	"testing"

	"github.com/google/uuid"
)

func prepareGroupTestFiles() error {

	// File with a good CSV file
	f, err := os.Create(file1)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(csvHeader)
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

	// File with no groups
	f, err = os.Create(file2)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(csvHeader)
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

	_, err = f.WriteString("name,descrsdiption,members,guac_config_protocol,guac_config_parameters\n")
	if err != nil {
		return err
	}
	_, err = f.WriteString(`"some","User","",,,` + "\n")
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

	_, err = f.WriteString(csvHeader)
	if err != nil {
		return err
	}
	_, err = f.WriteString(`"","test1","ron",,` + "\n")
	if err != nil {
		return err
	}
	f.Sync()

	return nil
}

func deleteGroupTestingFiles() {
	os.Remove(file1)
	os.Remove(file2)
	os.Remove(file3)
	os.Remove(file4)
}

func TestCsvCreateGroups(t *testing.T) {
	// Prepare test databases and echo testing server
	dbPath := uuid.New()
	e := testSetup(t, dbPath.String(), false)
	defer testCleanUp(dbPath.String())

	err := prepareGroupTestFiles()
	if err != nil {
		t.Fatal("error preparing CSV testing files")
	}
	defer deleteGroupTestingFiles()

	// Launch testing server
	go func() {
		e.Start(":50043")
	}()

	waitForTestServer(t, ":50043")
	const endpoint = "http://127.0.0.1:50043"
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
			cmd:            CsvCreateGroupsCmd(),
			args:           []string{serverFlag, endpoint, fileFlag, "/tmp/file1"},
			errorMessage:   "can't open CSV file",
			successMessage: "",
		},
		{
			name:           "create groups should be succesful",
			cmd:            CsvCreateGroupsCmd(),
			args:           []string{serverFlag, endpoint, fileFlag, file1},
			errorMessage:   "",
			successMessage: "devel: successfully created\nadmins: successfully created\nCreate from CSV finished!\n",
		},
		{
			name:           "group devel detail",
			cmd:            ListGroupCmd(),
			args:           []string{serverFlag, endpoint, "--gid", "1", "--json"},
			errorMessage:   "",
			successMessage: `{"gid":1,"name":"devel","description":"Developers","members":[{"uid":3,"username":"saul","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","jpeg_photo":"","manager":false,"readonly":false,"locked":false},{"uid":4,"username":"kim","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","jpeg_photo":"","manager":false,"readonly":false,"locked":false}],"guac_config_protocol":"","guac_config_parameters":""}` + "\n",
		},
		{
			name:           "repeat file, groups should be skipped",
			cmd:            CsvCreateGroupsCmd(),
			args:           []string{serverFlag, endpoint, fileFlag, file1},
			errorMessage:   "",
			successMessage: "devel: skipped, group already exists\nadmins: skipped, group already exists\nCreate from CSV finished!\n",
		},
		{
			name:           "file without groups",
			cmd:            CsvCreateGroupsCmd(),
			args:           []string{serverFlag, endpoint, fileFlag, file2},
			errorMessage:   "no groups where found in CSV file",
			successMessage: "",
		},
		{
			name:           "file with wrong header",
			cmd:            CsvCreateGroupsCmd(),
			args:           []string{serverFlag, endpoint, fileFlag, file3},
			errorMessage:   "wrong header",
			successMessage: "",
		},
		{
			name:           "file with groups that have several errors",
			cmd:            CsvCreateGroupsCmd(),
			args:           []string{serverFlag, endpoint, fileFlag, file4},
			errorMessage:   "",
			successMessage: ": skipped, required group name\nCreate from CSV finished!\n",
		},
	}

	for _, tc := range testCases {
		runTests(t, tc)
	}

}
