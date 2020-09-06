# gorepomod

A tool for managing Go modules in a git repository
with more than one Go module, where there are
dependencies between the modules.

This is a fancy version of

```
find ./ -name "go.mod" | xargs go mod {some operation}
```

Run it from a local git repository root.

It walks the repository's tree looking for Go modules
(i.e. `go.mod` files), loads and examines them all,
and does the following on each module _m_:

 - list
 
   Lists the modules and inter-repo dependencies.
   
 - tidy

   Tidy _m_'s go.mod file.

 - unpin {module}

   If _m_ depends on a _{repository}/{module}_,
   then _m_'s dependency on it will be replaced by
   a relative path to the in-repo module.

 - pin {module} {version}

   The opposite of 'unpin'.  Replacements are removed,
   and _m_'s dependency is pinned to a specific, previously
   tagged and released version of _{module}_.
   _{version}_ should be in semver form, e.g. `v1.2.3`.
