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

cd "$(dirname "$0")"
dir=$(pwd)
RULE="$1"
shift

for rule in semgrep/rules/$RULE/* ; do
	  echo semgrep --config "$rule" --autofix "$@"
	  semgrep --config "$rule" --autofix "$@"
done

# Comby fixes
comby -in-place -config "./comby/$RULE.toml" -f "$@"
# Cannot implement in just one pass
comby -in-place -config "./comby/$RULE.toml" -f "$@"

