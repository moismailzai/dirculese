[![Go Report Card](https://goreportcard.com/badge/github.com/moismailzai/dirculese)](https://goreportcard.com/report/github.com/moismailzai/dirculese) [![Build Status](https://travis-ci.org/moismailzai/dirculese.svg?branch=master)](https://travis-ci.org/moismailzai/dirculese) [![GoDoc](https://godoc.org/github.com/moismailzai/dirculese?status.svg)](https://godoc.org/github.com/moismailzai/dirculese) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
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
For now, only the ```ExtensionHandler``` exists, but there's plans for a ```DateHandler``` and a ```SizeHandler``` in the future.

### ExtensionHandler
ExtensionHandler iterates through all of the files in the directory that it is managing, and if any file has an extension that's listed in the ```Extensions``` array, that file will either be moved to the ```Target``` directory or deleted, depending on whether ```Delete``` is true or false. You can also add an empty entry to the ```Extensions``` array if you want to target files that do not have extensions.

## Contributing
Contributions are eagerly accepted.