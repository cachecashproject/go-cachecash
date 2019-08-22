#!/usr/bin/env bash

set -xeu -o pipefail

packr -v

# It's important to run sqlboiler with `-o .` so that `outputDirDepth` in the generated code is 0.  This makes it look
# for `sqlboiler.toml` in the same directory as the generated files, which lets us use different configuration files for
# different sets of models.  Otherwise, it looks only `outputDirDepth` levels above the models (here, it was looking in
# the repository's root).
pushd ../models/
rm -f *.go

shopt -s failglob

###
# this spawns a postgres container to apply fresh migrations to. kills it when bash dies
###
containerid=$(docker run -p 9999:5432 -d -e POSTGRES_DB=kvstore postgres:11)
while ! echo "select 1" | psql -U postgres -h localhost -p 9999 kvstore &>/dev/null
do
  sleep 1
done

end() {
  docker rm -f $containerid
}
trap end EXIT
###
# end postgres bit
###

sql-migrate up -config=../migrations/dbconfig.yml -env=postgres
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

# # For postgres (XXX: Will need paths adjusted)
# sql-migrate up -env=cache-development-pg
# sqlboiler -o ../models/ psql
