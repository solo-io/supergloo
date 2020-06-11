#!/bin/bash

find * -name '*.yaml' | while read -r f; do
  baseName=$(echo "$f" | sed -E 's|.yaml||')

  echo "// NOTE: Generated from $0
// DO NOT MODIFY

package operator_manifests

const $baseName = \`$(cat "$f")
\`
" > "$baseName".go
done
