### warning

this dir is for generating code into by the `ranger/ranger_test.go` test suite.
gitignore has been setup to avoid committing these generated files. These files
are compiled and generated on the fly by the tests.

If the build tag `rangertest` is not present in generated code and/or static
files in this directory, it may break other testing behavior that depends on
testing or compiling this directory. Please use the `rangertest` build tag for
all test code.
