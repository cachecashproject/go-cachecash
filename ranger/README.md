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
