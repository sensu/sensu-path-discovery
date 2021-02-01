[![Sensu Bonsai Asset](https://img.shields.io/badge/Bonsai-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/sensu/sensu-path-discovery)
![Go Test](https://github.com/sensu/sensu-path-discovery/workflows/Go%20Test/badge.svg)
![goreleaser](https://github.com/sensu/sensu-path-discovery/workflows/goreleaser/badge.svg)

# Sensu Path Discovery

## Table of Contents
- [Overview](#overview)
- [Usage examples](#usage-examples)
  - [Help output](#help-output)
  - [Examples](#examples)
  - [Paths files](#paths-files)
- [Configuration](#configuration)
  - [Asset registration](#asset-registration)
  - [Check definition](#check-definition)
- [Installation from source](#installation-from-source)
- [Additional notes](#additional-notes)
- [Contributing](#contributing)

## Overview

Discover file system paths and output a list of agent subscriptions. This plugin can
be used in combination with the [Sensu Entity Manager handler](https://github.com/sensu/sensu-entity-manager)
to automate Sensu Go agent subscription management.

## Usage examples

### Help output

```
Discover file system paths and output a list of agent subscriptions.

Usage:
  sensu-path-discovery [flags]
  sensu-path-discovery [command]

Available Commands:
  help        Help about any command
  version     Print the version number of this plugin

Flags:
  -f, --paths-file strings           The file location(s) for the mapping file (file path(s) or URL(s))
  -p, --subscription-prefix string   The agent subscription name prefix
  -t, --trusted-ca-file string       TLS CA certificate bundle in PEM format
  -i, --insecure-skip-verify         Skip TLS certificate verification (not recommended!)
  -h, --help                         help for sensu-path-discovery


Use "sensu-path-discovery [command] --help" for more information about a command.
```

### Examples

```
$ sensu-path-discovery --paths-file /etc/sensu/paths_to_subscriptions.json --paths-file http://artifacts.example.com/sensu/path-discovery.json
webapp
nginx
```

```
$ sensu-path-discovery -f /etc/sensu/paths_to_subscriptions.json -p ad:
ad:webapp
ad:nginx
```

### Paths files

This check makes use of operator provided file(s) that maps file system paths to Sensu
agent subscriptions.  The argument to the `--paths-file` option can be either a local file
path or http(s) URL. Multiple files are supported by either providing them as a comma
separated list argument to `--paths-file` or by specifying `--paths-file` multiple times.

The file needs to be in JSON in the following format:

```
[
	{
		"path": "/webapp",
		"subs": [
			"nginx",
			"webapp"
		]
	},
	{
		"path": "/var/lib/mysql",
		"subs": [
			"mysql"
		]
	},
	{
		"path": "/app/tomcat7",
		"subs": [
			"tomcat-common",
			"tomcat7"
		]
	},
	{
		"path": "/app/tomcat8",
		"subs": [
			"tomcat-common",
			"tomcat8"
		]
	}
]
```

Where `path` is a file system path and `subs` is an array of subscriptions to include for that path.

## Configuration

### Asset registration

[Sensu Assets][10] are the best way to make use of this plugin. If you're not using an asset, please
consider doing so! If you're using sensuctl 5.13 with Sensu Backend 5.13 or later, you can use the
following command to add the asset:

```
sensuctl asset add sensu/sensu-path-discovery
```

If you're using an earlier version of sensuctl, you can find the asset on the [Bonsai Asset Index](https://bonsai.sensu.io/assets/sensu/sensu-path-discovery).

### Check definition

```yml
---
type: CheckConfig
api_version: core/v2
metadata:
  name: sensu-path-discovery
  namespace: default
spec:
  command: >-
    sensu-path-discovery -p ad:
  subscriptions:
  - discovery
  runtime_assets:
  - sensu/sensu-path-discovery
  interval: 60
  publish: true
```

## Installation from source

The preferred way of installing and deploying this plugin is to use it as an Asset. If you would
like to compile and install the plugin from source or contribute to it, download the latest version
or create an executable from this source.

From the local path of the sensu-path-discovery repository:

```
go build
```

## Additional notes

## Contributing

For more information about contributing to this plugin, see [Contributing][1].

[1]: https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md
[10]: https://docs.sensu.io/sensu-go/latest/reference/assets/
