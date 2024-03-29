# hashbrowns

[![CircleCI](https://circleci.com/gh/sonatype-nexus-community/hashbrowns.svg?style=shield)](https://circleci.com/gh/sonatype-nexus-community/hashbrowns)
<a href="https://github.com/sonatype-nexus-community/hashbrowns/actions"><img src="https://github.com/sonatype-nexus-community/hashbrowns/workflows/nancy-gh-action/badge.svg" alt="nancy-gh-action"></img></a>

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

### Generating a shasum file

Depending on your operating system, you'll use something akin to `shasum` to get the sha1 and location of a file. A well formed `shasum` file looks like:

```
9987ca4f73d5ea0e534dfbf19238552df4de507e  main.go
2a72a07fbc9de22308d12a32f7d33504349e63c9  Makefile
```

`hashbrowns` is built to parse the output of `shasum` generated entries, and the important part here is `shasum` seems to put two spaces between the sha1 and the file name. If `hashbrowns` doesn't work for you, file an issue on our repo here, it is likely because the output of your `shasum` command is different.

### Nexus IQ Server Options

By default, assuming you have an out of the box Nexus IQ Server running, you can run `hashbrowns` like so:

`./hashbrowns fry --application public-application-id --path file-with-sha1-sums.txt`

It is STRONGLY suggested that you do not do this, and we will warn you on output if you are.

A more logical use of `hashbrowns` against Nexus IQ Server will look like so:

`./hashbrowns fry --application public-application-id --user nondefaultuser --token yourtoken --server-url http://adifferentserverurl:port --stage develop`

Options for stage are as follows:

`build, develop, stage-release, release`

By default `--stage` will be `develop`.

Successful submissions to Nexus IQ Server will result in either an OS exit of 0, meaning all is clear and a response akin to:

```
Wonderbar! No policy violations reported for this audit!
Report URL:  http://reportURL
```

Failed submissions will either indicate failure because of an issue with processing the request, or a policy violation. Both will exit with a code of 1, allowing you to fail your build in CI. Policy Violation failures will include a report URL where you can learn more about why you encountered a failure.

Policy violations will look like:

```
Hi, Hashbrowns here, you have some policy violations to clean up!
Report URL:  http://reportURL
```

Errors processing in Nexus IQ Server will look like:

```
Uh oh! There was an error with your request to Nexus IQ Server: <error>
```

## Development

`hashbrowns` is built with Golang, and specifically 1.14.2

To work on `hashbrowns`, fork/clone this repo, and ensure you have golang 1.14.2 installed, as well as Docker

We use a `Makefile` to consolidate build tasks, which by default is:

* Downloading dependencies
* Running `go test`
* Linting (uses Docker)
* Building

You can run `make` in the root of the repo, and those tasks will run.

`hashbrowns` was built using Cobra, and usage of Cobra is not super necessary, but sure doesn't hurt!

## Why Hashbrowns?

The program sends in hashes to Nexus IQ Server, and effectively looks for brown ones (bad ones). Punny, right?

## Installation

At current time you have a one option:

* Build from source
* Downloading a release from GitHub or Dockerhub coming soon!

## Contributing

We care a lot about making the world a safer place, and that's why we created `hashbrowns`. If you as well want to
speed up the pace of software development by working on this project, jump on in! Before you start work, create
a new issue, or comment on an existing issue, to let others know you are!

### Release Process

Follow the steps below to release a new version. You need to be part of the `deploy from circle ci` group for this to work.

1. Checkout/pull the latest `main` branch, and create a new tag with the desired semantic version and a helpful note:

       git tag -a v0.0.x -m "Helpful message in tag."

2. Push the tag up:

       git push origin v0.0.x

3. There is no step 3.

## The Fine Print

Remember:

It is worth noting that this is **NOT SUPPORTED** by Sonatype, and is a contribution of ours to the open source
community (read: you!)

* Use this contribution at the risk tolerance that you have
* Do NOT file Sonatype support tickets related to `ossindex-lib`
* DO file issues here on GitHub, so that the community can pitch in

Phew, that was easier than I thought. Last but not least of all - have fun!

## Getting help

Looking to contribute to our code but need some help? There's a few ways to get information:

* Chat with us on [Gitter](https://gitter.im/sonatype-nexus-community/hashbrowns)
