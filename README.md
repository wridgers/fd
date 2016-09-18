# fd

**f**in**d** text in files. Reads `.gitignore` and ignores VCS directories. Go routines are used for faster searching.

## Install ##

    $ go get -v github.com/wridgers/fd

## Update ##

    $ go get -u -v github.com/wridgers/fd

## Usage ##

    $ fd <term> <dir>

`term` accepts a regular expression and is required. `dir` is optional and defaults to `.`.

    $ fd -h
    Usage of fd:
      -i	search is case insensitive
      -v	invert search results
