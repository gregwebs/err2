#!/usr/bin/env bash
set -euo pipefail

RULE="$1"

mkdir -p result/$RULE/
cp case/$RULE/input.go result/$RULE/
# First build the test case to prove it is valid Go
output="result/$RULE/input.go"
go build "$output"

./apply-rule.sh $RULE "$output"

# Compare to golden
diff "$output" case/$RULE/golden.go
# Prove the golden case compiles
go build "$output"

# Cleanup
# If we fail before here, leave the file so it can be inspected
rm "$output"
