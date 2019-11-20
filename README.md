# gitclone

The `gitclone` command attempts to use git to clone the specified repo and then inspect the project so it can set it up.

## Supported project types

- [x] Go (modules)
- [x] NPM
- [x] Yarn

## Install

```sh
go get github.com/stevegood/gitclone && \
go install github.com/stevegood/gitclone
```

## Usage

`gitclone git@github.com:username/repo.git`

