/*
dirculese organizes your directories so you don't have to.
Usage:
	dirculese [flag]
The flags are:
	-silent
		suppress all messages to standard out and standard error (they are still logged)
	-config /full/path/to/your/config.json
		the full path to a dirculese configuration file
Before you can use dirculese, you will need to create a configuration file. By default, dirculese will try to load a
file called .dirculese.json in your home directory. Here's what a basic configuration file looks like:
	{
	  "Directories": [
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
This configuration only has a single directory with a single rule, but you can have as many directories and rules as you
want (dirculese will parse them in sequence).
If want to place your configuration file somewhere else, just call dirculese with the -config flag:
	dirculese -config /full/path/to/your/config.json
By default, dirculese is very verbose about what it's doing, but you can tell it to be silent with the -silent flag:
	dirculese -silent
Even when running silently, dirculese logs everything to dirculese.log which it saves to your home directory.
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
	"strconv"
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
	flag.BoolVar(&flagSilent, "silent", false, "suppress all messages to standard out and standard error (they are still logged)")
	flag.Parse()
}

// DirectoriesConfig is a simple struct that is used to map to the top-level array of directories in a dirculese JSON
// configuration file.
type DirectoriesConfig struct {
	Directories []DirectoryConfig
}

// DirectoryConfig is a simple struct that is used to map to a single directory in a dirculese JSON configuration file.
type DirectoryConfig struct {
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

// Directory is the basic type of a managed directory. Directories are managed based on the Rule items in the
// Directory.rules slice, which are executed sequentially by Directory.Ruler(). The Directory.path string should be an
// existing, accessible directory, which is validated by calling Directory.CheckPath()
type Directory struct {
	rules []Rule
	path  string
}

// Rule defines a single criteria for managing a directory. Rule.source is a pointer to a Directory representation of
// the source directory and Rule.target is a pointer to a Directory representation of the target directory. Any files in
// the source directory that match the rule's criteria will be moved into the target directory, unless Rule.delete is
// true, in which case the files will be deleted instead. Rule.handler is the name of the handler function that should
// be used to execute the rule's logic, and is parsed by Rule.Handler().
type Rule struct {
	source     *Directory
	target     *Directory
	delete     bool
	handler    string
	extensions []string
	sizeMax    int
	sizeMin    int
	dateMax    int
	dateMin    int
}

// CheckPath tests to see if a directory's d.path points to an existing directory on the filesystem.
func (d *Directory) CheckPath() (err error) {
	var fileInfo os.FileInfo
	if d.path == "" {
		return errors.New("empty paths are not valid")
	}
	fileInfo, err = os.Stat(d.path)
	if err != nil {
		return errors.New(err.Error())
	}
	if !fileInfo.IsDir() {
		return errors.New(d.path + " is not a directory")
	}
	return
}

// Contents returns the contents of a directory's d.path.
func (d *Directory) Contents() (contents []os.FileInfo, err error) {
	contents, err = ioutil.ReadDir(d.path)
	if err != nil {
		err = errors.New(err.Error())
	}
	return
}

// Ruler sequentially executes the individuals rules in a directory's d.rules slice.
func (d *Directory) Ruler() (err error) {
	for _, element := range d.rules {
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

	// make sure the path we're going to be moving items into exists and is accessible
	err = r.target.CheckPath()
	if err != nil {
		return errors.New(err.Error())
	}

	// make a map of all the extensions so lookups are easier later
	fileExtensions := make(map[string]string)
	for _, extension := range r.extensions {
		fileExtensions[extension] = extension
	}

	// get a list of all the items in the directory we're managing
	files, err := r.source.Contents()
	if err != nil {
		return errors.New(err.Error())
	}

	// for each item
	for _, f := range files {
		// if it's a file
		if !f.IsDir() {
			var message string
			fileExtension := strings.TrimLeft(filepath.Ext(f.Name()), ".")
			// and if this file's extension is in the map we created earlier
			if _, extensionExists := fileExtensions[fileExtension]; extensionExists {
				// if the delete flag is set, delete the file
				if r.delete {
					err = os.Remove(r.source.path + string(os.PathSeparator) + f.Name())
					message = "Deleted the file " + f.Name() + " in the path " + r.source.path + "."
				} else {
					// otherwise, check to see if a file by this name exists in the destination directory, if not, move it
					if _, e := os.Stat(r.target.path + string(os.PathSeparator) + f.Name()); os.IsNotExist(e) {
						err = os.Rename(r.source.path+string(os.PathSeparator)+f.Name(), r.target.path+string(os.PathSeparator)+f.Name())
						message = "Moved the file " + f.Name() + " from the path " + r.source.path + " to " + r.target.path + "."
					} else {
						// if so, try appending numbers to the end of the filename and check if a file by the new name exists
						// give up after 9998 tries (entirely arbitrary)
						for i := 0; i < 9999; i++ {
							newFileName := strings.TrimRight(f.Name(), filepath.Ext(f.Name())) + strconv.Itoa(i) + filepath.Ext(f.Name())
							if _, e := os.Stat(r.target.path + string(os.PathSeparator) + newFileName); os.IsNotExist(e) {
								err = os.Rename(r.source.path+string(os.PathSeparator)+f.Name(), r.target.path+string(os.PathSeparator)+newFileName)
								message = "Moved the file " + f.Name() + " from the path " + r.source.path + " to " + r.target.path + " (renamed to " + newFileName + ") because a file with the same name already exists there."
								break
							}
							if i == 9998 {
								message = "Didn't move the file " + f.Name() + " from the path " + r.source.path + " to " + r.target.path + " because a file with the same name already exists there."
							}
						}
					}
				}
				if err != nil {
					return errors.New(err.Error())
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

// GetConfigStruct loads a JSON file from the given path on the filesystem and maps its contents to a DirectoriesConfig
// struct.
func GetConfigStruct(path string) (conf DirectoriesConfig, err error) {
	conf = DirectoriesConfig{}
	confFile, err := os.Open(path)
	if err != nil {
		return
	}
	decoder := json.NewDecoder(confFile)
	decoder.Decode(&conf)
	if len(conf.Directories) < 1 {
		err = errors.New("your configuration file should include at least one directory")
	}
	return
}

// GetDirectories creates an array of Directories (including the rules associated with each one) based on the contents
// of a DirectoriesConfig struct.
func GetDirectories(config DirectoriesConfig) (directories []Directory) {
	for _, directoryConf := range config.Directories {
		d := Directory{}
		d.path = directoryConf.Path
		for _, ruleConf := range directoryConf.Rules {
			rule := Rule{}
			targetDirectory := Directory{path: ruleConf.Target}
			rule.source = &d
			rule.target = &targetDirectory
			rule.delete = ruleConf.Delete
			rule.handler = ruleConf.Handler
			rule.extensions = ruleConf.Extensions
			rule.sizeMax = ruleConf.SizeMax
			rule.sizeMin = ruleConf.SizeMin
			rule.dateMax = ruleConf.DateMax
			rule.dateMin = ruleConf.DateMin
			d.rules = append(d.rules, rule)
		}
		directories = append(directories, d)
	}
	return
}

// GetSampleConfig generates a sample dirculese configuration file.
func GetSampleConfig() (config string) {
	config = `{"Directories":[{"Path":"/path/to/a/source/directory/that/you/want/to/keep/organized/with/dirculese/rules","Rules":[{"Target":"/path/to/a/destination/directory/where/items/matching/your/rule/will/be/moved","Delete":false,"Handler":"ExtensionHandler","Extensions":["png"],"SizeMax":0,"SizeMin":0,"DateMax":0,"DateMin":0}]}]}`
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
		return errors.New("the JSON in your configuration file cannot be validated")
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
		message += "' is incorrectly formatted: "
		message += err.Error()
		message += ". Here's what a valid Dirculese configuration file looks like: "
		message += GetSampleConfig()
		message += " See https://github.com/moismailzai/dirculese for more information."
		logError.Fatalln(message)
	}

	// use the configuration struct to build directories and rules
	directories := GetDirectories(configStruct)

	for _, directory := range directories {
		err := directory.Ruler()
		if err != nil {
			logError.Fatalln(err.Error())
		}
	}

	os.Exit(0)
}
