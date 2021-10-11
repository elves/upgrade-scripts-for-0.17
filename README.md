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

If the assignment form contains multiple variables, and some of them already
exist while others don't, it rewrites it to a `var` form that declares the
non-existing variables and a `set` form:

```
a = foo
a b = lorem ipsum
# becomes
var a = foo
var b; set a b = lorem ipsum
```

## What this doesn't do

This program does not handle any other changes introduced in 0.17.

## How to use

**This program is still in development. Don't use it yet.**
