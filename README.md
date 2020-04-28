# hashbrowns

[![CircleCI](https://circleci.com/gh/sonatype-nexus-community/hashbrowns.svg?style=svg)](https://circleci.com/gh/sonatype-nexus-community/hashbrowns)

Hashbrowns is a utility for scanning sha1 sums akin to:

```
9987ca4f73d5ea0e534dfbf19238552df4de507e  main.go
```

With Sonatype's Nexus IQ Server.

## Usage

```
$ hashbrowns 
Actual usage of this tool is used with the fry command. Please see hashbrowns fry --help for more information.

Usage:
  hashbrowns [command]

Available Commands:
  fry         Submit list of sha1s to Nexus IQ Server
  help        Help about any command

Flags:
  -v, -- count   Set log level, higher is more verbose
  -h, --help     help for hashbrowns

Use "hashbrowns [command] --help" for more information about a command.
```

```
$ hashbrowns fry --help
Provided a path to a file with sha1's and locations, this command will submit them to Nexus IQ Server.

This can be used to audit generic environments for matches to known hashes that do not meet your org's policy.

Usage:
  hashbrowns fry [flags]

Flags:
      --application string   Specify application ID for request
  -h, --help                 help for fry
      --max-retries int      Specify maximum number of tries to poll Nexus IQ Server (default 300)
      --path string          Path to file with sha1s
      --server-url string    Specify Nexus IQ Server URL (default "http://localhost:8070")
      --stage string         Specify stage for application (default "develop")
      --token string         Specify Nexus IQ token/password for request (default "admin123")
      --user string          Specify Nexus IQ username for request (default "admin")

Global Flags:
  -v, -- count   Set log level, higher is more verbose
```
