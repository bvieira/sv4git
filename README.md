<p align="center">
  <h1 align="center">sv4git</h1>
  <p align="center">semantic version for git</p>
  <p align="center">
    <a href="https://github.com/bvieira/sv4git/releases/latest"><img alt="Release" src="https://img.shields.io/github/release/bvieira/sv4git.svg?style=for-the-badge"></a>
    <a href="https://github.com/bvieira/sv4git/stargazers"><img alt="GitHub stars" src="https://img.shields.io/github/stars/bvieira/sv4git?style=for-the-badge"></a>
    <a href="/LICENSE"><img alt="Software License" src="https://img.shields.io/badge/license-MIT-informational.svg?style=for-the-badge"></a>
    <a href="https://github.com/bvieira/sv4git/actions?workflow=ci"><img alt="GitHub Actions" src="https://img.shields.io/github/workflow/status/bvieira/sv4git/ci?style=for-the-badge"></a>
    <a href="https://goreportcard.com/report/github.com/bvieira/sv4git"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/bvieira/sv4git?style=for-the-badge"></a>
    <a href="https://conventionalcommits.org"><img alt="Software License" src="https://img.shields.io/badge/Conventional%20Commits-1.0.0-informational.svg?style=for-the-badge"></a>
  </p>
</p>

## Getting Started

### Installing

- Download the latest release and add the binary on your path
- Optional: Set `SV4GIT_HOME` to define user configs, check [config](#config) for more information.

### Config

There are 3 config levels when using sv4git: [default](#default), [user](#user), [repository](#repository). All 3 are merged using the follow priority: **repository > user > default**.

To see current config, run:

```bash
git sv cfg show
```

#### Configuration types

##### Default

To check what is the default configuration, run:

```bash
git sv cfg default
```

##### User

To configure define `SV4GIT_HOME` environment variable, eg.:

```bash
SV4GIT_HOME=/home/myuser/.sv4git # myuser is just an example
```

And define the `config.yml` inside it, eg:

```bash
.sv4git
└── config.yml
```

##### Repository

Create a `.sv4git.yml` on the root of your repository, eg.: [.sv4git.yml](.sv4git.yml)

#### Configuration format

```yml
version: "1.0" #config version

versioning: # versioning bump
    update-major: [] # commit types used to bump major
    update-minor: # commit types used to bump minor
        - feat
    update-patch: # commit types used to bump patch
        - build
        - ci
        - chore
        - docs
        - fix
        - perf
        - refactor
        - style
        - test
    # when type is not present on update rules and is unknown (not mapped on commit message types), 
    # if ignore-unknown=false bump patch, if ignore-unknown=true do not bump version
    ignore-unknown: false 

tag:
    pattern: '%d.%d.%d' # pattern used to create git tag

release-notes:
    headers: # headers names for relase notes markdown, to disable a section, just remove the header line
        breaking-change: Breaking Changes
        feat: Features
        fix: Bug Fixes

branches: # git branches config
    prefix: ([a-z]+\/)? # prefix used on branch name, should be a regex group
    suffix: (-.*)? # suffix used on branch name, should be a regex group
    disable-issue: false # set true if there is no need to recover issue id from branch name
    skip: # list of branch names ignored on commit message validation
        - master
        - main
        - developer

commit-message:
    types: # supported commit types
        - build
        - ci
        - chore
        - docs
        - feat
        - fix
        - perf
        - refactor
        - revert
        - style
        - test
    scope:
        # define supported scopes, if blank, scope will not be validated, if not, only scope listed will be valid.
        # don't forget to add "" on your list if you need to define scopes and keep it optional
        values: [] 
    footer:
        issue:
            key: jira # name used to define an issue on footer metadata
            key-synonyms: # supported variations for footer metadata
                - Jira
                - JIRA
            use-hash: false # if false, use :<space> separator, if true, use <space># separator
    issue:
        regex: '[A-Z]+-[0-9]+' # regex for issue id
```

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

use `--help` or `-h` to get usage information, don't forget that some commands have unique options too

```bash
# sv help
git-sv -h

# sv release-notes command help
git-sv rn -h
```

##### Available commands

| Variable                     | description                                                   | has options or subcommands |
| ---------------------------- | ------------------------------------------------------------- | :------------------------: |
| config, cfg                  | show config information                                       |     :heavy_check_mark:     |
| current-version, cv          | get last released version from git                            |            :x:             |
| next-version, nv             | generate the next version based on git commit messages        |            :x:             |
| commit-log, cl               | list all commit logs according to range as jsons              |     :heavy_check_mark:     |
| commit-notes, cn             | generate a commit notes according to range                    |     :heavy_check_mark:     |
| release-notes, rn            | generate release notes                                        |     :heavy_check_mark:     |
| changelog, cgl               | generate changelog                                            |     :heavy_check_mark:     |
| tag, tg                      | generate tag with version based on git commit messages        |            :x:             |
| commit, cmt                  | execute git commit with convetional commit message helper     |            :x:             |
| validate-commit-message, vcm | use as prepare-commit-message hook to validate commit message |     :heavy_check_mark:     |
| help, h                      | shows a list of commands or help for one command              |            :x:             |

##### Use range

Commands like `commit-log` and `commit-notes` has a range option. Supported range types are: `tag`, `date` and `hash`.

By default, it's used [--date=short](https://git-scm.com/docs/git-log#Documentation/git-log.txt---dateltformatgt) at `git log`, all dates returned from it will be in `YYYY-MM-DD` format.

Range `tag` will use `git describe` to get the last tag available if `start` is empty, the others types won't use the existing tags, it's recommended to always use a start limit in a old repository with a lot of commits. This behavior was maintained to not break the retrocompatibility.

Range `date` use git log `--since` and `--until`, it's possible to use all supported formats from [git log](https://git-scm.com/docs/git-log#Documentation/git-log.txt---sinceltdategt), if `end` is in `YYYY-MM-DD` format, `sv` will add a day on git log command to make the end date inclusive.

Range `tag` and `hash` are used on git log [revision range](https://git-scm.com/docs/git-log#Documentation/git-log.txt-ltrevisionrangegt). If `end` is empty, `HEAD` will be used instead.

```bash
# get commit log as json using a inclusive range
git-sv commit-log --range hash --start 7ea9306~1 --end c444318

# return all commits after last tag
git-sv commit-log --range tag
```

##### Use validate-commit-message as prepare-commit-msg hook

Configure your .git/hooks/prepare-commit-msg

```bash
#!/bin/sh

COMMIT_MSG_FILE=$1
COMMIT_SOURCE=$2
SHA1=$3

git sv vcm --path "$(pwd)" --file $COMMIT_MSG_FILE --source $COMMIT_SOURCE
```

tip: you can configure a directory as your global git templates using the command below, check [git config docs](https://git-scm.com/docs/git-config#Documentation/git-config.txt-inittemplateDir) for more information!

```bash
git config --global init.templatedir '<YOUR TEMPLATE DIR>'
```

## Development

### Makefile

run `make` to get the list of available actions

```bash
make
```

#### Make configs

| Variable   | description            |
| ---------- | ---------------------- |
| BUILDOS    | build OS               |
| BUILDARCH  | build arch             |
| ECHOFLAGS  | flags used on echo     |
| BUILDENVS  | var envs used on build |
| BUILDFLAGS | flags used on build    |

| Parameters | description                         |
| ---------- | ----------------------------------- |
| args       | parameters that will be used on run |

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
