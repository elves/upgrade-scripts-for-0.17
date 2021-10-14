# Elvish script upgrader for 0.17

## What this does

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

## What this doesn't do

This program does not handle any other changes introduced in 0.17.

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
