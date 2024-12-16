# addlicense

The program ensures source code files have copyright license headers
by scanning directory patterns recursively.

It modifies all source files in place and avoids adding a license header
to any file that already has one.

addlicense requires go 1.16 or later.

## install

    go install github.com/google/addlicense@latest

## usage

    addlicense [flags] pattern [pattern ...]

    -c      copyright holder (default "Google LLC")
    -check  check only mode: verify presence of license headers and exit with non-zero code if missing
    -f      license file
    -ignore file patterns to ignore, for example: -ignore **/*.go -ignore vendor/**
    -l      license type: apache, bsd, mit, mpl (default "apache")
    -s      Include SPDX identifier in license header. Set -s=only to only include SPDX identifier.
    -v      verbose mode: print the name of the files that are modified
    -y      copyright year(s) (default "2022")

The pattern argument can be provided multiple times, and may also refer
to single files.  Directories are processed recursively.

For example, to run addlicense across everything in the current directory and
all subdirectories:

    addlicense .

The `-ignore` flag can use any pattern [supported by
doublestar](https://github.com/bmatcuk/doublestar#patterns).

## Running in a Docker Container

The simplest way to get the addlicense docker image is to pull from GitHub
Container Registry:

```bash
docker pull ghcr.io/google/addlicense:latest
```

Alternately, you can build it from source yourself:

```bash
docker build -t ghcr.io/google/addlicense .
```

Once you have the image, you can test that it works by running:

```bash
docker run -it ghcr.io/google/addlicense -h
```

Finally, to run it, mount the directory you want to scan to `/src` and pass the
appropriate addlicense flags:

```bash
docker run -it -v ${PWD}:/src ghcr.io/google/addlicense -c "Google LLC" *.go
```

## license

Apache 2.0

This is not an official Google product.
