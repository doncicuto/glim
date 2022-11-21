package cmd

import (
	"os"
	"testing"

	"github.com/google/uuid"
)

const file1 = "/tmp/file1.csv"
const file2 = "/tmp/file2.csv"
const file3 = "/tmp/file3.csv"
const file4 = "/tmp/file4.csv"
const csvHeader = "name,description,members,guac_config_protocol,guac_config_parameters\n"

func prepareGroupGuacamoleTestFiles() error {

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
	_, err = f.WriteString(`"programmers","Developers","saul,kim","vnc","host=localhost"` + "\n")
	if err != nil {
		return err
	}
	_, err = f.WriteString(`"managers","Administratos","kim","ssh","host=localhost"` + "\n")
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

	_, err = f.WriteString("name,descriptddion,membdsders,guac_condsdsfig_protocol,guac_config_parameters\n")
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

	_, err = f.WriteString(csvHeader)
	if err != nil {
		return err
	}
	_, err = f.WriteString(`"","test1","ron","",""` + "\n")
	if err != nil {
		return err
	}
	_, err = f.WriteString(`"test4","test1","kim","","port=22"` + "\n")
	if err != nil {
		return err
	}
	f.Sync()

	return nil
}

func deleteGroupGuacamoleTestingFiles() {
	os.Remove(file1)
	os.Remove(file2)
	os.Remove(file3)
	os.Remove(file4)
}

func TestCsvCreateGuacamoleGroups(t *testing.T) {
	// Prepare test databases and echo testing server
	dbPath := uuid.New()
	guacamoleEnabled := true
	e := testSetup(t, dbPath.String(), guacamoleEnabled) // guacamole <- true
	defer testCleanUp(dbPath.String())

	err := prepareGroupGuacamoleTestFiles()
	if err != nil {
		t.Fatal("error preparing CSV testing files")
	}
	defer deleteGroupGuacamoleTestingFiles()

	// Launch testing server
	go func() {
		e.Start(":50033")
	}()

	waitForTestServer(t, ":50033")
	const endpoint = "http://127.0.0.1:50033"
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
			successMessage: "programmers: successfully created\nmanagers: successfully created\nCreate from CSV finished!\n",
		},
		{
			name:           "group devel detail",
			cmd:            ListGroupCmd(),
			args:           []string{serverFlag, endpoint, "--gid", "1", "--json"},
			errorMessage:   "",
			successMessage: `{"gid":1,"name":"programmers","description":"Developers","members":[{"uid":3,"username":"saul","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","jpeg_photo":"","manager":false,"readonly":false,"locked":false},{"uid":4,"username":"kim","name":"","firstname":"","lastname":"","email":"","ssh_public_key":"","jpeg_photo":"","manager":false,"readonly":false,"locked":false}],"guac_config_protocol":"vnc","guac_config_parameters":"host=localhost"}` + "\n",
		},
		{
			name:           "repeat file, groups should be skipped",
			cmd:            CsvCreateGroupsCmd(),
			args:           []string{serverFlag, endpoint, fileFlag, file1},
			errorMessage:   "",
			successMessage: "programmers: skipped, group already exists\nmanagers: skipped, group already exists\nCreate from CSV finished!\n",
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
			successMessage: ": skipped, required group name\ntest4: skipped, Apache Guacamole config protocol is required\nCreate from CSV finished!\n",
		},
	}

	for _, tc := range testCases {
		runTests(t, tc)
	}

}
