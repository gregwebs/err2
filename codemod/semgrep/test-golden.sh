#!/usr/bin/env bash
set -euo pipefail
set -x

RULE="$1"

mkdir -p result/$RULE/
cp case/$RULE/input.go result/$RULE/
# First build the test case to prove it is valid Go
go build result/$RULE/input.go
# when autofixing, rules must be applied individually
for rule in rules/$1/* ; do
	  echo semgrep --config "$rule" "result/$1/input.go" --autofix
	  semgrep --config "$rule" "result/$1/input.go" --autofix
done
# Fix imports
goimports -w result/$RULE/input.go
# Compare to golden
diff result/$RULE/input.go case/$RULE/golden.go
# Prove the golden case compiles
go build result/$RULE/input.go
# Cleanup
rm result/$RULE/input.go 
