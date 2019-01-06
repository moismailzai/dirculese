[![Go Report Card](https://goreportcard.com/badge/github.com/moismailzai/dirculese)](https://goreportcard.com/report/github.com/moismailzai/dirculese) [![Build Status](https://travis-ci.org/moismailzai/dirculese.svg?branch=master)](https://travis-ci.org/moismailzai/dirculese) [![codecov](https://codecov.io/gh/moismailzai/dirculese/branch/master/graph/badge.svg)](https://codecov.io/gh/moismailzai/dirculese) [![GoDoc](https://godoc.org/github.com/moismailzai/dirculese?status.svg)](https://godoc.org/github.com/moismailzai/dirculese) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![Dirculese logo which depicts a standard folder icon with a muscular arm attached on the left ](dirculese.png "hero of song and story")
Dirculese organizes your directories so you don't have to.

## Installation

```
go get github.com/moismailzai/dirculese
```

## Usage
Before you can use dirculese, you will need to create a configuration file. By default, dirculese will try to load a file called ```.dirculese.json``` in your home directory. Here's what a basic configuration file looks like:

```

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
          "PrefixDelimiters": [
            "__"
          ],
          "SuffixDelimiters": [
            "--"
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
```
This simple configuration only has a single directory with a single rule, but you can have as many directories and rules as you want (dirculese will parse them in sequence).

If want to place your configuration file somewhere else, just call dirculese with the ```-config``` flag:

```
dirculese -config /full/path/to/your/config.json
```

By default, dirculese is very verbose about what it's doing, but you can tell it to be silent with the ```-silent``` flag:

```
dirculese -silent
```

Even when running silently, dirculese logs everything to ```dirculese.log``` which it saves in your home directory.

Dirculese returns an exit code of ```0``` if everything went well and an exit code of ```1``` if something went wrong.

## Dirculese handlers
For now, only the ```ExtensionHandler```, ```PrefixHandler```, and ```SuffixHandler``` exist, but there's plans for a ```DateHandler``` and a ```SizeHandler``` in the future.

### ExtensionHandler
ExtensionHandler iterates through all of the files in the directory that it is managing, and if any file has an extension that's listed in the ```Extensions``` array, that file will either be moved to the ```Target``` directory or deleted, depending on whether ```Delete``` is true or false. You can also add an empty entry to the ```Extensions``` array if you want to target files that do not have extensions.

### PrefixHandler
PrefixHandler iterates through all of the files in the directory that it is managing and targets any file whose name portion (excluding extension) includes a substring in the ```PrefixDelimiters``` array. Matching files are either deleted (if ```Delete``` is true) or moved into a subdirectory of ```Target```. Subdirectories are named using the portion of the file's name that **precedes** the prefix delimiter and are automatically created if they don't already exist.

For example, consider the below file listing:

```
pre1__test1.txt
pre1__test2.txt
pre1__test3.txt
pre2--test1.txt
pre2--test2.txt
pre2--test3.txt
pre3++test1.txt
```

And the below rules:

```
"Rules": [
    {
      "Target": "/path/to/a/target/directory",
      "Delete": false,
      "Handler": "PrefixHandler",
      "Extensions": [],
      "PrefixDelimiters": [
        "__",
        "--"
      ],
      "SuffixDelimiters": [],
      "SizeMax": 0,
      "SizeMin": 0,
      "DateMax": 0,
      "DateMin": 0
    },{
      "Target": "/path/to/a/target/directory",
      "Delete": true,
      "Handler": "PrefixHandler",
      "Extensions": [],
      "PrefixDelimiters": [
        "++",
      ],
      "SuffixDelimiters": [],
      "SizeMax": 0,
      "SizeMin": 0,
      "DateMax": 0,
      "DateMin": 0
    }
  ]
```

In this case, there would be two new subdirectories created in ```/path/to/a/target/directory```: ```pre1``` and ```pre2```. 

The contents of ```pre1``` would be:

```
pre1__test1.txt
pre1__test2.txt
pre1__test3.txt
```

The contents of ```pre2``` would be:

```
pre2--test1.txt
pre2--test2.txt
pre2--test3.txt
```

Besides ```pre1``` and ```pre2```, ```/path/to/a/target/directory``` would be empty because the last file, ```pre3++test1.txt``` would have been deleted (based on the second rule).

### SuffixHandler
SuffixHandler iterates through all of the files in the directory that it is managing and targets any file whose name portion (excluding extension) includes a substring in the ```SuffixDelimiters``` array. Matching files are either deleted (if ```Delete``` is true) or moved into a subdirectory of ```Target```. Subdirectories are named using the portion of the file's name that **follows** the prefix delimiter and are automatically created if they don't already exist.

See the examples from PrefixHandler.

## Contributing
Contributions are eagerly accepted.

