# sv4git

Semantic version for git

## Getting Started

### Installing

Comming soon...

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
