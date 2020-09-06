# gorepomod

Manages the Go modules in a git repository.

Install:
```
go install github.com/monopole/gorepomod
```

It handles tasks one might otherwise attempt with

```
find ./ -name "go.mod" | xargs {some hack involving go mod and git}
```

Run it from a git repository root.

It walks the repository, reading `go.mod` files and
building a model of Go modules and intra-repo module
dependencies (if any).

## Usage

#### `gorepomod tidy`

Creates a change with mechanical updates
to `go.mod` and `go.sum` files.

#### `gorepomod list`

Lists modules and intra-repo dependencies.

A good way to get module names for use in
the following commands.

#### `gorepomod unpin {module}`

Creates a change to edit `go.mod` files.

For each module _m_ in the repository,
if _m_ depends on a _{module}_,
then _m_'s dependency on it will be replaced by
a relative path to the in-repo module.

#### `gorepomod pin {module} [{version}]`

Creates a change to edit `go.mod` files.

The opposite of `unpin`.  The change removes
replacements and pins _m_ to a specific, previously
tagged and released version of _{module}_.

_{version}_ should be in semver form, e.g. `v1.2.3`.

If _{version}_ is omitted, _m_ will be pinned
to the most recent version of _{module}_.

#### `gorepomod release {module} [major|minor|patch]`

Computes a new version for the module and "releases" it,
meaning it creates a tag then pushes the tag, and a release
branch, upstream.

If the existing version is _v1.2.7_, the new version
will be _v2.0.0_, _v1.3.0_ or _v1.2.8_, depending on
the value of the 2nd command argument.  The 2nd argument
defaults to `patch`.  This establishes values for
_{major}_, _{minor}_ and _{patch}_ in what follows.

This command looks for a branch named

> _release-{module}/-v{major}.{minor}_

If it doesn't exist, it creates it and pushes it upstream.

Then the command creates a new tag in the form

> _{module}/v{major}.{minor}.{patch}_

The command pushes this tag upstream, which might
trigger cloud activity.

#### `gorepomod unrelease {module}`

This undoes the work of `release`, by deleting the
most recent tag locally and upstream.

This must be done immediately after a release. 
If there's a chance someone (or some auto) already imported the module
at the given tag, then don't do this, because it will
confuse module caches.  Do a new patch release instead.

