# sv4git

Semantic version for git

## Getting Started

### Installing

Download the latest release and add the binary on your path

## Running

run `git-sv` to get the list of available parameters

```bash
git-sv
```

### Run as git command

if `git-sv` is configured on your path, you can use it like a git command

```bash
git sv
git sv current-version
git sv next-version
```

### Usage

```bash
NAME:
   sv - semantic version for git

USAGE:
   git-sv [global options] command [command options] [arguments...]

VERSION:
   1.0.0

COMMANDS:
   current-version, cv  get last released version from git
   next-version, nv     generate the next version based on git commit messages
   commit-log, cl       list all commit logs since last version as jsons
   release-notes, rn    generate release notes
   tag, tg              generate tag with version based on git commit messages
   help, h              Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version

```

## Development

### Makefile

run `make` to get the list of available actions

```bash
make
```

#### Make configs

| Variable | description|
| --------- | ----------|
| BUILDOS | build OS |
| BUILDARCH | build arch |
| ECHOFLAGS | flags used on echo |
| BUILDENVS | var envs used on build |
| BUILDFLAGS | flags used on build |

| Parameters | description|
| --------- | ----------|
| args | parameters that will be used on run |

```bash
#variables
BUILDOS="linux" BUILDARCH="amd64" make build

#parameters
make run args="-h"
```

### Build

```bash
make build
```

the binary will be created on `bin/$BUILDOS_$BUILDARCH/git-sv`

### Tests

```bash
make test
```

### Run

```bash
#without args
make run

#with args
make run args="-h"
```
