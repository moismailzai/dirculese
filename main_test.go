package main

import (
	"io/ioutil"
	"os"
	"os/user"
	"runtime"
	"strings"
	"testing"
)

var (
	configStruct = DirectoriesConfig{
		Directories: []DirectoryConfig{{
			Path: "/path/to/a/source/directory/that/you/want/to/keep/organized/with/dirculese/rules",
			Rules: []RuleConfig{{
				Target:     "/path/to/a/destination/directory/where/items/matching/your/rule/will/be/moved",
				Delete:     false,
				Handler:    "ExtensionHandler",
				Extensions: []string{"png"},
				SizeMax:    0,
				SizeMin:    0,
				DateMax:    0,
				DateMin:    0,
			}},
		}},
	}
	directories = []Directory{{
		path: "/path/to/a/source/directory/that/you/want/to/keep/organized/with/dirculese/rules",
		rules: []Rule{{
			target:     &Directory{path: "/path/to/a/destination/directory/where/items/matching/your/rule/will/be/moved"},
			delete:     false,
			handler:    "ExtensionHandler",
			extensions: []string{"png"},
			sizeMax:    0,
			sizeMin:    0,
			dateMax:    0,
			dateMin:    0,
		}},
	}}
	sampleConfig = `{"Directories":[{"Path":"/path/to/a/source/directory/that/you/want/to/keep/organized/with/dirculese/rules","Rules":[{"Target":"/path/to/a/destination/directory/where/items/matching/your/rule/will/be/moved","Delete":false,"Handler":"ExtensionHandler","Extensions":["png"],"SizeMax":0,"SizeMin":0,"DateMax":0,"DateMin":0}]}]}`
)

func init() {
	directories[0].rules[0].source = &directories[0]
}

func TestDirectory_CheckPath(t *testing.T) {
	_, dir, _, _ := runtime.Caller(0)
	dir = strings.TrimRight(dir, "main_tes.go")

	testDirectoryPass := Directory{path: dir + "testing"}
	testDirectoryFail := Directory{path: dir + "PATH-DOES-NOT-EXIST"}

	var want error
	got := testDirectoryPass.CheckPath()

	if want != got {
		t.Errorf("Valid directory failed check: "+testDirectoryPass.path+". Got '%v', want '%v'", got, want)
	}

	got = testDirectoryFail.CheckPath()

	if want == got {
		t.Errorf("Invalid directory passed check: "+testDirectoryFail.path+". Got '%v'", got)
	}

}

func TestDirectory_Contents(t *testing.T) {
	_, dir, _, _ := runtime.Caller(0)
	dir = strings.TrimRight(dir, "main_tes.go")

	testDirectory := Directory{path: dir + "testing"}

	want, errWant := ioutil.ReadDir(dir + "testing")
	got, errGot := testDirectory.Contents()

	if len(want) != len(got) {
		t.Errorf("Didn't get the right number of items from "+testDirectory.path+". Got '%v', want '%v'", len(got), len(want))
	}

	if errWant != errGot {
		t.Errorf("Couldn't get the contents of directory "+testDirectory.path+". Got '%v', want '%v'", got, want)
	}
}

func TestDirectory_Ruler(t *testing.T) {
	want := "you need to specify at least one extension"

	testDirectory := Directory{}
	testDirectory.rules = append(testDirectory.rules, Rule{})
	testDirectory.rules[0].handler = "ExtensionHandler"

	err := testDirectory.Ruler()
	got := err.Error()

	if want != got {
		t.Errorf("The correct handler was not run. Got '%v', want '%v'", got, want)
	}
}

func TestGetConfigFilePath(t *testing.T) {
	userHome, _ := GetUserHome()

	want := userHome + string(os.PathSeparator) + DefaultConfigFile
	got, _ := GetConfigFilePath()
	if got != want {
		t.Errorf("Configuration file path mismatch. Got '%v', want '%v'", got, want)
	}

	flagConfig = "TEST"
	want = flagConfig
	got, _ = GetConfigFilePath()
	if got != want {
		t.Errorf("Configuration file path mismatch. Got '%v', want '%v'", got, want)
	}
}

func TestGetConfigStruct(t *testing.T) {

	want := configStruct
	got, _ := GetConfigStruct("./testing/dirculese.test.json")

	if got.Directories[0].Rules[0].Target != want.Directories[0].Rules[0].Target {
		t.Errorf("Mismatch in Target. Got '%v', want '%v'", got.Directories[0].Rules[0].Target, want.Directories[0].Rules[0].Target)
	}
	if got.Directories[0].Rules[0].Delete != want.Directories[0].Rules[0].Delete {
		t.Errorf("Mismatch in Delete. Got '%v', want '%v'", got.Directories[0].Rules[0].Delete, want.Directories[0].Rules[0].Delete)
	}
	if got.Directories[0].Rules[0].Handler != want.Directories[0].Rules[0].Handler {
		t.Errorf("Mismatch in Handler. Got '%v', want '%v'", got.Directories[0].Rules[0].Handler, want.Directories[0].Rules[0].Handler)
	}
	if got.Directories[0].Rules[0].Extensions[0] != want.Directories[0].Rules[0].Extensions[0] {
		t.Errorf("Mismatch in Extensions[0]. Got '%v', want '%v'", got.Directories[0].Rules[0].Extensions[0], want.Directories[0].Rules[0].Extensions[0])
	}
	if got.Directories[0].Rules[0].SizeMax != want.Directories[0].Rules[0].SizeMax {
		t.Errorf("Mismatch in SizeMax. Got '%v', want '%v'", got.Directories[0].Rules[0].SizeMax, want.Directories[0].Rules[0].SizeMax)
	}
	if got.Directories[0].Rules[0].SizeMin != want.Directories[0].Rules[0].SizeMin {
		t.Errorf("Mismatch in SizeMin. Got '%v', want '%v'", got.Directories[0].Rules[0].SizeMin, want.Directories[0].Rules[0].SizeMin)
	}
	if got.Directories[0].Rules[0].DateMax != want.Directories[0].Rules[0].DateMax {
		t.Errorf("Mismatch in DateMax. Got '%v', want '%v'", got.Directories[0].Rules[0].DateMax, want.Directories[0].Rules[0].DateMax)
	}
	if got.Directories[0].Rules[0].DateMin != want.Directories[0].Rules[0].DateMin {
		t.Errorf("Mismatch in DateMin. Got '%v', want '%v'", got.Directories[0].Rules[0].DateMin, want.Directories[0].Rules[0].DateMin)
	}
}

func TestGetDirectories(t *testing.T) {

	want := directories
	got := GetDirectories(configStruct)

	if got[0].rules[0].target.path != want[0].rules[0].target.path {
		t.Errorf("Mismatch in path. Got '%v', want '%v'", got[0].rules[0].target.path, want[0].rules[0].target.path)
	}
	if got[0].rules[0].delete != want[0].rules[0].delete {
		t.Errorf("Mismatch in delete. Got '%v', want '%v'", got[0].rules[0].delete, want[0].rules[0].delete)
	}
	if got[0].rules[0].handler != want[0].rules[0].handler {
		t.Errorf("Mismatch in handler. Got '%v', want '%v'", got[0].rules[0].handler, want[0].rules[0].handler)
	}
	if got[0].rules[0].extensions[0] != want[0].rules[0].extensions[0] {
		t.Errorf("Mismatch in extensions[0]. Got '%v', want '%v'", got[0].rules[0].extensions[0], want[0].rules[0].extensions[0])
	}
	if got[0].rules[0].sizeMax != want[0].rules[0].sizeMax {
		t.Errorf("Mismatch in sizeMax. Got '%v', want '%v'", got[0].rules[0].sizeMax, want[0].rules[0].sizeMax)
	}
	if got[0].rules[0].sizeMin != want[0].rules[0].sizeMin {
		t.Errorf("Mismatch in sizeMin. Got '%v', want '%v'", got[0].rules[0].sizeMin, want[0].rules[0].sizeMin)
	}
	if got[0].rules[0].dateMax != want[0].rules[0].dateMax {
		t.Errorf("Mismatch in dateMax. Got '%v', want '%v'", got[0].rules[0].dateMax, want[0].rules[0].dateMax)
	}
	if got[0].rules[0].dateMin != want[0].rules[0].dateMin {
		t.Errorf("Mismatch in dateMin. Got '%v', want '%v'", got[0].rules[0].dateMin, want[0].rules[0].dateMin)
	}
}

func TestGetSampleConfig(t *testing.T) {
	want := sampleConfig
	got := GetSampleConfig()
	if got != want {
		t.Errorf("Config mismatch. Got '%v', want '%v'", got, want)
	}
}

func TestGetUserHome(t *testing.T) {
	currentUser, _ := user.Current()
	want := currentUser.HomeDir
	got, _ := GetUserHome()
	if got != want {
		t.Errorf("User home mismatch. Got '%v', want '%v'", got, want)
	}
}

func TestRule_ExtensionHandler(t *testing.T) {
	// get path to the directory the test is running in
	_, dir, _, _ := runtime.Caller(0)
	dir = strings.TrimRight(dir, "main_tes.go")

	// create Directory and Rule objects for the test
	testDirectory := Directory{path: dir + "testing"}
	testDirectory.rules = []Rule{
		{
			source:     &testDirectory,
			target:     &Directory{path: dir + "testing" + string(os.PathSeparator) + "doc"},
			handler:    "ExtensionHandler",
			extensions: []string{"doc", "pdf"},
		}, {
			source:     &testDirectory,
			target:     &Directory{path: dir + "testing" + string(os.PathSeparator) + "img"},
			handler:    "ExtensionHandler",
			extensions: []string{"png", "jpg"},
		}, {
			source:     &testDirectory,
			target:     &Directory{path: dir + "testing" + string(os.PathSeparator) + "noext"},
			handler:    "ExtensionHandler",
			extensions: []string{""},
		}, {
			source:     &testDirectory,
			handler:    "ExtensionHandler",
			delete:     true,
			extensions: []string{"del"},
		},
	}

	// create mock files and directories inside the testing directory
	mockFiles := []string{"test.png", "test.jpg", "test.doc", "test.pdf", "test", "test.del"}
	mockDirectories := []string{"doc", "img", "noext"}
	for _, mockDirectory := range mockDirectories {
		os.RemoveAll(dir + "testing" + string(os.PathSeparator) + mockDirectory)
		os.MkdirAll(dir+"testing"+string(os.PathSeparator)+mockDirectory, 0777)
	}
	for _, mockFile := range mockFiles {
		_, err := os.OpenFile(dir+"testing"+string(os.PathSeparator)+mockFile, os.O_RDONLY|os.O_CREATE, 0777)
		if err != nil {
			t.Error("Error while creating mock files for this test: " + err.Error())
		}
	}

	// run the first test, expecting no errors and for all mock files to have been moved out of the testing directory
	// and into the appropriate mock directory (but dirculese.test.json should still be present). Using Ruler() means
	// that every Rule's .ExtensionHandler method will be run in sequence.
	var want error
	got := testDirectory.Ruler()

	if want != got {
		t.Errorf("Something went wrong, ExtensionHandler returned an error. Got '%v', want '%v'", got, want)
	}

	// now add more mock files to the testing directory
	for _, file := range mockFiles {
		_, err := os.OpenFile(dir+"testing"+string(os.PathSeparator)+file, os.O_RDONLY|os.O_CREATE, 0777)
		if err != nil {
			t.Error("Error while creating mockFiles for this test: " + err.Error())
		}
	}

	// run the second test, again expecting no errors and for all mock files to have been moved out of the test
	// directory and into the appropriate mock directory (and dirculese.test.json just still be present). also expecting
	// the subdirectories to have two of each mock files, with the second file having a 0 appended to its name
	got = testDirectory.Ruler()
	if want != got {
		t.Errorf("Something went wrong, an ExtensionHandler returned an error. Got '%v', want '%v'", got, want)
	}

	// now build a table to verify the test results
	type directoryTest struct {
		directory string
		want      string
	}
	directoryTestTable := []directoryTest{
		{
			directory: testDirectory.path,
			want:      "dirculese.test.json",
		}, {
			directory: testDirectory.rules[0].target.path,
			want:      "test.doc,test.pdf,test0.doc,test0.pdf",
		}, {
			directory: testDirectory.rules[1].target.path,
			want:      "test.jpg,test.png,test0.jpg,test0.png",
		}, {
			directory: testDirectory.rules[2].target.path,
			want:      "test,test0",
		},
	}

	// verify results
	for _, d := range directoryTestTable {
		filesString := ""
		directoryTest := Directory{path: d.directory}
		fileInfos, err := directoryTest.Contents()
		if err != nil {
			t.Error("Error while getting the contents of" + d.directory + ": " + err.Error())
		}
		for _, fileInfo := range fileInfos {
			if !fileInfo.IsDir() {
				filesString += fileInfo.Name() + ","
			}
		}
		got := strings.TrimRight(filesString, ",")
		if d.want != got {
			t.Errorf("Incorrect filelist in "+testDirectory.path+". Got '%v', want '%v'", got, d.want)
		}
	}

	// remove all mock files and directories that were created for this test
	for _, targetDirectory := range mockDirectories {
		os.RemoveAll(dir + "testing" + string(os.PathSeparator) + targetDirectory)
	}

}

func TestRule_Handler(t *testing.T) {
	want := "you need to specify at least one extension"

	testRule := Rule{}
	testRule.handler = "ExtensionHandler"

	err := testRule.Handler()
	got := err.Error()

	if want != got {
		t.Errorf("The correct handler was not run. Got '%v', want '%v'", got, want)
	}
}

func TestValidateConfigFile(t *testing.T) {
	_, dir, _, _ := runtime.Caller(0)
	dir = strings.TrimRight(dir, "main_tes.go")

	var want error
	got := ValidateConfigFile(dir + "testing/dirculese.test.json")
	if got != want {
		t.Errorf("Couldn't validate config file. Got '%v', want '%v'", got, want)
	}
}
