# gorepomod

A tool to do bulk, interrelated edits to `go.mod`
files in one git repository.

Only useful if you have more than one _Go_ module
in a repository, and they depend on each other.

Usage:

```
gorepomod unpin {dependency}
gorepomod pin {dependency} {version}
gorepomod tidy
```

e.g.

```
gorepomod unpin kyaml --doIt
gorepomod pin kyaml v0.7.0 --doIt
```

This program must be run from a local git repository root.
The program walks the repository's tree looking for Go
modules (i.e. `go.mod` files), and performs one of the
following operations on each module _m_:

 - tidy

   Tidy _m_'s go.mod file.

 - unpin

   If _m_ depends on a _{repository}/{dependency}_,
   then _m_'s dependency on it will be replaced by
   a relative path to the in-repo module.

 - pin {version}

   The opposite of 'unpin'.  Replacements are removed,
   and _m_'s dependency is pinned to a specific, previously
   tagged and released version of _{dependency}_.
   _{version}_ should be in semver form, e.g. `v1.2.3`.
