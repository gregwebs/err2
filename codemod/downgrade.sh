#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 ]] ; then
	echo "  $0 [files]"
	echo ""
	echo "expects at least 1 argument"
	echo ""
	exit 0
fi

exec apply-rule.sh downgrade "$@"
