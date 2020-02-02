# sv4git

Semantic version for git

## Getting Started

### Installing

download the latest release and add the binary on your path

### Config

you can config using the environment variables

| Variable | description | default |
| --------- | ----------| ----------|
|MAJOR_VERSION_TYPES|types used to bump major version||
|MINOR_VERSION_TYPES|types used to bump minor version|feat|
|PATCH_VERSION_TYPES|types used to bump patch version|build,ci,docs,fix,perf,refactor,style,test|
|INCLUDE_UNKNOWN_TYPE_AS_PATCH|force patch bump on unknown type|true|
|BRAKING_CHANGE_PREFIXES|list of prefixes that will be used to identify a breaking change|BREAKING CHANGE:,BREAKING CHANGES:|
|ISSUEID_PREFIXES|list of prefixes that will be used to identify an issue id|jira:,JIRA:,Jira:|
|TAG_PATTERN|tag version pattern|%d.%d.%d|
|RELEASE_NOTES_TAGS|release notes headers for each visible type|fix:Bug Fixes,feat:Features|

### Running

run `git-sv` to get the list of available parameters

```bash
git-sv
```

#### Run as git command

if `git-sv` is configured on your path, you can use it like a git command

```bash
git sv
git sv current-version
git sv next-version
```

#### Usage

use `--help` or `-h` to get usage information, dont forget that some commands have unique options too

```bash
# sv help
git-sv -h

# sv release-notes command help
git-sv rn -h
```

##### Available commands

| Variable | description | has options |
| --------- | ---------- | :----------: |
| current-version, cv | get last released version from git | :x: |
| next-version, nv | generate the next version based on git commit messages | :x: |
| commit-log, cl | list all commit logs since last version as jsons | :heavy_check_mark: |
| release-notes, rn | generate release notes | :heavy_check_mark: |
| changelog, cgl | generate changelog | :heavy_check_mark: |
| tag, tg | generate tag with version based on git commit messages | :x: |
| help, h | Shows a list of commands or help for one command | :x: |

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
