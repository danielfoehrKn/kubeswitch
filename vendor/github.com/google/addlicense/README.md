# addlicense

The program ensures source code files have copyright license headers
by scanning directory patterns recursively.

It modifies all source files in place and avoids adding a license header
to any file that already has one.

addlicense requires go 1.16 or later.

## install

    go get -u github.com/google/addlicense

## usage

    addlicense [flags] pattern [pattern ...]

    -c copyright holder (defaults to "Google LLC")
    -f custom license file (no default)
    -l license type: apache, bsd, mit, mpl (defaults to "apache")
    -y year (defaults to current year)
    -check check only mode: verify presence of license headers and exit with non-zero code if missing
    -ignore file patterns to ignore, for example: -ignore **/*.go -ignore vendor/**

The pattern argument can be provided multiple times, and may also refer
to single files.

The `-ignore` flag can use any pattern [supported by
doublestar](https://github.com/bmatcuk/doublestar#patterns).

## Running in a Docker Container

- Clone the repository using `git clone https://github.com/google/addlicense.git`
- Build your docker container
```bash
docker build -t google/addlicense .
```

- Test the image
```bash
docker run -it google/addlicense -h
```

- Usage example
```bash
docker run -v ${PWD}:/src -it google/addlicense -c "Google LLC" *.go
```

## license

Apache 2.0

This is not an official Google product.
