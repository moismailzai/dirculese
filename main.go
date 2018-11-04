/*
dirculese organizes your directories so you don't have to.
Usage:
	dirculese [flag]
The flags are:
	-silent
		suppresses all messages to standard out and standard error (they are still logged)
	-config /full/path/to/your/config.json
		the full path to a dirculese configuration file
Before you can use dirculese, you will need to create a configuration file. By default, dirculese will try to load a
file called .dirculese.json in your home directory. Here's what a basic configuration file looks like:
	{
	  "Folders": [
		{
		  "Path": "/path/to/a/source/directory/that/you/want/to/keep/organized/with/dirculese/rules",
		  "Rules": [
			{
			  "Target": "/path/to/a/destination/directory/where/items/matching/your/rule/will/be/moved",
			  "Delete": false,
			  "Handler": "ExtensionHandler",
			  "Extensions": [
				"png"
			  ],
			  "SizeMax": 0,
			  "SizeMin": 0,
			  "DateMax": 0,
			  "DateMin": 0
			}
		  ]
		}
	  ]
	}
This simple configuration only has a single directory with a single rule, but you can have as many directories and rules
as you want (dirculese will parse them in sequence).
If want to place your configuration file somewhere else, just call dirculese with the -config flag:
	dirculese -config /full/path/to/your/config.json
By default, dirculese is very verbose about what it's doing, but you can tell it to be silent with the -silent flag:
	dirculese -silent
Even when running silently, dirculese logs everything to dirculese.log which it saves in your home directory.
Dirculese returns an exit code of 0 if everything went well and an exit code of 1 if something went wrong.
*/
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// DefaultConfigFile is the name of the file in the user's home directory that dirculese will use for its configuration.
// DefaultLogFile is the name of the file in the user's home directory that dirculese will log to.
const (
	DefaultConfigFile = ".dirculese.json"
	DefaultLogFile    = "dirculese.log"
)

var (
	flagConfig  string
	flagSilent  bool
	logStandard *log.Logger
	logError    *log.Logger
)

func init() {
	flag.StringVar(&flagConfig, "config", "", "the full path to a dirculese configuration file")
	flag.BoolVar(&flagSilent, "silent", false, "suppresses all messages to standard out and standard error (they are still logged)")
	flag.Parse()
}

// FoldersConfig is a simple struct that is used to map to the top-level array of folders in a dirculese JSON
// configuration file.
type FoldersConfig struct {
	Folders []FolderConfig
}

// FolderConfig is a simple struct that is used to map to a single folder in a dirculese JSON configuration file.
type FolderConfig struct {
	Rules []RuleConfig
	Path  string
}

// RuleConfig is a simple struct that is used to map to a single rule in a dirculese JSON configuration file.
type RuleConfig struct {
	Target     string
	Delete     bool
	Handler    string
	Extensions []string
	SizeMax    int
	SizeMin    int
	DateMax    int
	DateMin    int
}

// Folder is the basic type of a managed folder. Folders are managed based on the Rule items in the Folder.rules slice,
// which are executed sequentially by Folder.Ruler(). The Folder.path string should be a valid, accessible directory,
// which is validated by calling Folder.CheckPath()
type Folder struct {
	rules []Rule
	path  string
}

// Rule defines the criteria for managing a folder. Rule.source is a pointer to a Folder representation of the source
// directory and Rule.target is a pointer to a Folder representation of the target directory. Any files in the source
// directory that match the rule's criteria will be moved into the target directory, unless Rule.delete is true, in
// which case the files will be deleted instead. Rule.handler is the name of the handler function that should be used to
// execute the rule's logic, and is parsed by Rule.Handler().
type Rule struct {
	source     *Folder
	target     *Folder
	delete     bool
	handler    string
	extensions []string
	sizeMax    int
	sizeMin    int
	dateMax    int
	dateMin    int
}

// CheckPath tests to see if a folder's f.path points to a valid directory on the filesystem.
func (f *Folder) CheckPath() (err error) {
	var fileInfo os.FileInfo
	if f.path == "" {
		return errors.New("empty paths are not valid")
	}
	fileInfo, err = os.Stat(f.path)
	if err != nil {
		return errors.New(err.Error())
	}
	if !fileInfo.IsDir() {
		return errors.New(f.path + " is not a directory")
	}
	return
}

// Contents returns the contents of a folder's f.path.
func (f *Folder) Contents() (contents []os.FileInfo, err error) {
	contents, err = ioutil.ReadDir(f.path)
	return
}

// Ruler sequentially executes the individuals rules in a folder's f.rules slice.
func (f *Folder) Ruler() (err error) {
	for _, element := range f.rules {
		err = element.Handler()
		if err != nil {
			return errors.New(err.Error())
		}
	}
	return
}

// Handler reads a rule's r.handler property and maps it to a predefined handler. This allows rules that are defined in
// text configuration files to be easily mapped to handler methods.
func (r *Rule) Handler() (err error) {
	switch r.handler {
	case "ExtensionHandler":
		err = r.ExtensionHandler()
	default:
		err = errors.New("unrecognized handler")
	}
	return
}

// ExtensionHandler iterates through all of the files in a rule's r.source directory, and if any file has an extension
// that's listed in the r.extensions slice, it is either moved into the r.target directory or deleted, depending on the
// boolean state of r.delete
func (r *Rule) ExtensionHandler() (err error) {
	if len(r.extensions) < 1 {
		return errors.New("you need to specify at least one extension")
	}
	if r.delete != true {
		if r.target == new(Folder) {
			return errors.New("you need to specify a target directory")
		}
	}
	err = r.target.CheckPath()
	if err != nil {
		return errors.New(err.Error())
	}
	fileExtensions := make(map[string]string)
	for _, extension := range r.extensions {
		fileExtensions[extension] = extension
	}
	files, err := r.source.Contents()
	if err != nil {
		return errors.New(err.Error())
	}
	for _, f := range files {
		if !f.IsDir() {
			thisExtension := strings.TrimLeft(filepath.Ext(f.Name()), ".")

			if _, exists := fileExtensions[thisExtension]; exists {
				if r.delete {
					err = os.Remove(r.source.path + string(os.PathSeparator) + f.Name())
				} else {
					err = os.Rename(r.source.path+string(os.PathSeparator)+f.Name(), r.target.path+string(os.PathSeparator)+f.Name())
				}
				if err != nil {
					return errors.New(err.Error())
				}
				var message string
				if r.delete {
					message = "Deleted the file " + f.Name() + " in the path " + r.source.path + "."
				} else {
					message = "Moved the file " + f.Name() + " from the path " + r.source.path + " to " + r.target.path + "."
				}
				logStandard.Println(message)
			}
		}
	}
	return
}

// GetConfigFilePath returns the full path to the user's dirculese configuration file. If a -config flag was specified,
// its argument will be used verbatim. Otherwise, the path to the user's home directory will be prepended to the OS's
// path separator and the constant DefaultConfigFile.
func GetConfigFilePath() (path string, err error) {
	path = ""
	if flagConfig == "" {
		path, err = GetUserHome()
		if err != nil {
			return
		}
		path += string(os.PathSeparator) + DefaultConfigFile
	} else {
		path = flagConfig
	}
	return
}

// GetConfigStruct loads a JSON file from the given path on the filesystem and maps its contents to a FoldersConfig
// struct.
func GetConfigStruct(path string) (conf FoldersConfig, err error) {
	conf = FoldersConfig{}
	confFile, err := os.Open(path)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(confFile)
	decoder.Decode(&conf)
	return
}

// GetFolders creates an array of Folders (including the rules associated with each one) based on the contents of a
// FoldersConfig struct.
func GetFolders(c FoldersConfig) (folders []Folder) {
	for _, folderConf := range c.Folders {
		f := Folder{}
		f.path = folderConf.Path
		for _, ruleConf := range folderConf.Rules {
			rule := Rule{}
			targetFolder := Folder{path: ruleConf.Target}
			rule.source = &f
			rule.target = &targetFolder
			rule.delete = ruleConf.Delete
			rule.handler = ruleConf.Handler
			rule.extensions = ruleConf.Extensions
			rule.sizeMax = ruleConf.SizeMax
			rule.sizeMin = ruleConf.SizeMin
			rule.dateMax = ruleConf.DateMax
			rule.dateMin = ruleConf.DateMin
			f.rules = append(f.rules, rule)
		}
		folders = append(folders, f)
	}
	return
}

// GetSampleConfig generates a sample dirculese configuration file.
func GetSampleConfig() (config string) {
	config = `{"Folders":[{"Path":"/path/to/a/source/directory/that/you/want/to/keep/organized/with/dirculese/rules","Rules":[{"Target":"/path/to/a/destination/directory/where/items/matching/your/rule/will/be/moved","Delete":false,"Handler":"ExtensionHandler","Extensions":["png"],"SizeMax":0,"SizeMin":0,"DateMax":0,"DateMin":0}]}]}`
	return
}

// GetUserHome returns the runtime user's home directory.
func GetUserHome() (home string, err error) {
	currentUser, _ := user.Current()
	home = currentUser.HomeDir
	if home == "" {
		err = errors.New("can't find your home directory (try using the -config flag with the full path to your config file)")
	}
	return
}

// ValidateConfigFile checks to see if the configuration file at the provided path exists on the filesystem and is in
// JSON format.
func ValidateConfigFile(path string) (err error) {
	configFile := new(os.File)
	configFile, err = os.Open(path)
	if err != nil {
		return errors.New(err.Error())
	}
	defer configFile.Close()
	confContents, err := ioutil.ReadAll(configFile)
	if err != nil {
		return errors.New(err.Error())
	}
	if !json.Valid(confContents) {
		return errors.New(err.Error())
	}
	return
}

func main() {

	// setup logging
	userHome, _ := GetUserHome()
	logFile, err := os.OpenFile(userHome+string(os.PathSeparator)+DefaultLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)

	if err != nil {
		log.Fatalln("Failed to open log file '"+DefaultLogFile+"': ", err.Error())
	}

	defer logFile.Close()

	// suppress standard output if the -silent flag was used
	if flagSilent {
		logStandard = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)
		logError = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		logStandard = log.New(io.MultiWriter(logFile, os.Stdout), "", log.Ldate|log.Ltime|log.Lshortfile)
		logError = log.New(io.MultiWriter(logFile, os.Stderr), "", log.Ldate|log.Ltime|log.Lshortfile)
	}

	// load the configuration file
	configFilePath, err := GetConfigFilePath()

	if err != nil {
		message := "Whoops: "
		message += err.Error()
		message += "."
		logError.Fatalln(message)
	}

	// validate that the configuration file exists and is in json format
	err = ValidateConfigFile(configFilePath)

	if err != nil {
		message := "Whoops, there is a problem loading or parsing your configuration file '"
		message += configFilePath
		message += "': "
		message += err.Error()
		message += ". Here's what a valid Dirculese configuration file looks like: "
		message += GetSampleConfig()
		message += " See https://github.com/moismailzai/dirculese for more information."
		logError.Fatalln(message)
	}

	// map the configuration file to a configuration struct
	configStruct, err := GetConfigStruct(configFilePath)

	if err != nil {
		message := "Whoops, your configuration file '"
		message += configFilePath
		message += "' is incorrectly formatted. "
		message += "Here's what a valid Dirculese configuration file looks like: "
		message += GetSampleConfig()
		message += " See https://github.com/moismailzai/dirculese for more information."
		logError.Fatalln(message)
	}

	// use the configuration struct to build folders and rules
	folders := GetFolders(configStruct)

	for _, folder := range folders {
		err := folder.Ruler()
		if err != nil {
			logError.Fatalln(err.Error())
		}
	}

	os.Exit(0)
}
