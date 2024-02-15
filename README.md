# Frizbee

[![Coverage Status](https://coveralls.io/repos/github/stacklok/frizbee/badge.svg?branch=main)](https://coveralls.io/github/stacklok/frizbee?branch=main) | [![License: Apache 2.0](https://img.shields.io/badge/License-Apache2.0-brightgreen.svg)](https://opensource.org/licenses/Apache-2.0) | [![](https://dcbadge.vercel.app/api/server/RkzVuTp3WK?logo=discord&label=Discord&color=5865&style=flat)](https://discord.gg/RkzVuTp3WK)

---

Frizbee is a tool you may throw a tag at and it comes back with a checksum.

It's a command-line tool designed to provide checksums for GitHub Actions
and container images based on tags.

It also includes a set of libraries for working with tags and checksums.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
  - [GitHub Actions](#github-actions)
  - [Container Images](#container-images)
- [Configuration](#configuration)
- [Commands](#commands)
- [Autocompletion](#autocompletion)
- [Contributing](#contributing)
- [License](#license)

## Installation

To install Frizbee, you can use the following methods:

```bash
# Using Go
go get -u github.com/stacklok/frizbee
go install github.com/stacklok/frizbee

# Using Homebrew
brew install stacklok/tap/frizbee

# Using winget
winget install stacklok.frizbee
```

## Usage

### GitHub Actions

Frizbee can be used to generate checksums for GitHub Actions. This is useful
for verifying that the contents of a GitHub Action have not changed.

To quickly replace the GitHub Action references for your project, you can use
the `ghactions` command:

```bash
frizbee ghactions -d path/to/your/repo/.github/workflows/
```

This will write all the replacements to the files in the directory provided.

Note that this command will only replace the `uses` field of the GitHub Action
references.

Note that this command supports dry-run mode, which will print the replacements
to stdout instead of writing them to the files.

It also supports exiting with a non-zero exit code if any replacements are found. 
This is handy for CI/CD pipelines.

If you want to generate the replacement for a single GitHub Action, you can use
the `ghactions one` command:

```bash
frizbee ghactions one metal-toolbox/container-push/.github/workflows/container-push.yml@main
```

This is useful if you're developing and want to quickly test the replacement.

### Container Images

Frizbee can be used to generate checksums for container images. This is useful
for verifying that the contents of a container image have not changed.

To get the digest for a single image tag, you can use the `containerimage one` command:

```bash
frizbee containerimage one quay.io/stacklok/frizbee:latest
```

This will print the image refrence with the digest for the image tag provided.


## Contributing

We welcome contributions to Frizbee. Please see our [Contributing](./CONTRIBUTING.md) guide for more information.

## License

Frizbee is licensed under the [Apache 2.0 License](./LICENSE).
