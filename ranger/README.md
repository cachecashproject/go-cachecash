# Stable byte serialisation a-la protobuf

## Code Design

Note: this is aspirational.

The config file describes the desired native language types and their serialised behaviours.

This is parsed into a data structure that provides an abstraction that can render a native language type and the
required public (de)serialisation code for it.

For golang this is comprised of:

* The struct definition
* The Marshal method
* The MarshalTo method
* The Unmarshal Method
* The UnMarshalFrom method
* The Size method

Each type is expected to self delimit - no exterior containers are supplied. If a given type isn't of well defined size,
then the serialisation code for that type should include its own length prefix tag.

Additionally though, to allow for clean and safe deserialisation, each type also creates additional helper methods:

* MinimumSize - the smallest size that this type is permitted to be deserialised into. This is useful as it allows
                rolling up the requirements for structs and the like.

The abstract data structure has methods on it to emit the above methods, but also, to help write out those methods,
it can be asked to write out valid default value - this is distinct from a zero value, because pointer types will be
given an instance of the type they point at.

* DefaultValueFor

## On tests

Tests generate and then run generated tests on the generated code. This code is
generated in `./ranger/testdata/pkg`. If you wish to re-run the tests without a
generation step (to test code you hacked up), run:

```
$ go test -race -v ./ranger/testdata/pkg -count 1
```

To fuzz:

```
$ make fuzz FUZZ=ranger/testdata/pkg
```

After running the first generation (AKA, run this first):

```
$ go test -race -v ./ranger
```
