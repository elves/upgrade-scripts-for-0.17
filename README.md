# Elvish script upgrader for 0.17

## What this does

### Rewriting legacy assignment forms

This program rewrites legacy assignment forms to equivalent `var` or `set`
forms:

```sh
a = foo
# becomes:
var a = foo

# but when $a is already defined...
a = foo
# becomes:
set a = foo
```

If the assignment form contains a mix of existing and new variables, it is
rewritten to a `var` form that declares the non-existing variables and a `set`
form:

```sh
a = foo
a b = lorem ipsum
# becomes
var a = foo
var b; set a b = lorem ipsum
```

The version of the `set` command in 0.15.x and 0.16.x contained a bug where it
could also create variables. This program also rewrites such uses of `set` by
declaring those variables with `var` first:

```sh
set a = foo
set a b = lorem ipsum
# becomes
var a; set a = foo
var b; set a b = lorem ipsum
```

### Rewriting legacy lambda syntax

This program also rewrites legacy lambda syntax to the new syntax, moving
arguments and options within `[...]` before `{` to within `|...|` after `{`.

```sh
x = [a b &k=v]{ ... }
# becomes
x = {|a b &k=v| ... }
```

The new lambda syntax is supported since 0.17.0. If your script still needs to
support older versions, you can turn off lambda rewrite with `-lambda=false`.

## What this doesn't do

This program does not handle any other changes introduced in 0.17.

## Known limitations

In the legacy assignment form, the RHS may refer to the just declared variable,
such as:

```sh
m = [&x={ put $m }]
```

This will get rewritten to the following, which doesn't work since the `var`
form now evaluates the RHS before declaring the variable:

```sh
var m = [&x={ put $m }]
```

You will need to manually rewrite this to:

```sh
var m
set m = [&x={ put $m }]
```

## How to use

Build this program:

```sh
go install github.com/elves/upgrade-scripts-for-0.17@main
```

This will install the program to ~/go/bin by default. Add ~/go/bin to your PATH,
or copy the program to somewhere already on PATH.

Command line invocation works like `gofmt` - there are 3 modes:

```sh
upgrade-scripts-for-0.17 # no arguments; rewrite stdin to stdout
upgrade-scripts-for-0.17 a.elv b.elv # rewrites script to stdout
upgrade-scripts-for-0.17 -w a.elv b.elv # rewrites script in place
```

If you're invoking it from Elvish, use the following to rewrite all Elvish
scripts in the current directory recursively:

```sh
upgrade-scripts-for-0.17 -w **[type:regular].elv
```

Remember to back up the files, or make sure that they are in version control,
just in case this program has bugs and renders your scripts unusable.
