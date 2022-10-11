#!/usr/bin/env bash
set -euo pipefail

if ! command -v semgrep >/dev/null ; then
	echo "semgrep must be installed and in the PATH" >&2
	exit 1
fi

if ! command -v comby >/dev/null ; then
	echo "comby must be installed and in the PATH" >&2
	exit 1
fi

if ! command -v goimports >/dev/null ; then
	echo "goimports must be installed and in the PATH" >&2
	exit 1
fi

if [[ $# -lt 2 ]] ; then
	echo "  apply-rule.sh RULE file [files]"
	echo ""
	echo "     RULE  upgrade or downgrade"
	echo ""
	echo "expects at least 2 arguments"
	echo ""
	exit 0
fi

pushd "$(dirname "$0")"
dir=$(pwd)
popd

RULE="$1"
shift

for rule in "$dir/semgrep/rules/$RULE/"* ; do
	  echo semgrep -l go --config "$rule" --autofix "$@"
	  semgrep -l go --config "$rule" --autofix "$@"
done

# Comby fixes
comby -in-place -config "$dir/comby/$RULE.toml" -f "$@"
# Cannot implement in just one pass
comby -in-place -config "$dir/comby/$RULE.toml" -f "$@"

# Fix imports
goimports -w "$@"
gofmt -w "$@"
