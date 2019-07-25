#!/usr/bin/env bash

set -eu -o pipefail
#shopt -s failglob

packr -v

# It's important to run sqlboiler with `-o .` so that `outputDirDepth` in the generated code is 0.  This makes it look
# for `sqlboiler.toml` in the same directory as the generated files, which lets us use different configuration files for
# different sets of models.  Otherwise, it looks only `outputDirDepth` levels above the models (here, it was looking in
# the repository's root).
pushd ../models/
rm -f *.go
sql-migrate up -config=../migrations/dbconfig.yml -env=development
sqlboiler -c ../migrations/sqlboiler.toml -o . psql
popd

# Add build tag to generated tests.
for SRCFILE in ../models/*_test.go; do
    ed -s "${SRCFILE}" <<EOF
1i
// +build sqlboiler_test

.
w
q
EOF

done
