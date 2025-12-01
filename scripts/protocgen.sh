#!/usr/bin/env bash
set -e

echo "Generating Protocol Buffer code..."
cd proto

# Generate slinky protos
proto_dirs=$(find ./slinky -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
    if grep go_package $file &> /dev/null ; then
      buf generate --template buf.gen.gogo.yaml $file
    fi
  done
done

# Generate sidecar protos
proto_dirs=$(find ./sidecar -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
    if grep go_package $file &> /dev/null ; then
      buf generate --template buf.gen.gogo.yaml $file
    fi
  done
done

cd ..

# move proto files to the right places
if [ -d "github.com/skip-mev/slinky" ]; then
  cp -r github.com/skip-mev/slinky/* ./
fi

rm -rf github.com

# go mod tidy --compat=1.20
