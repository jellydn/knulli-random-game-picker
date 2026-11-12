#!/bin/sh
set -eu

base=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
export PICKER_STATE_DIR="$base/state"
umask 077
mkdir -p "$PICKER_STATE_DIR"
# This entry never opens SDL or changes the normal launcher's startup behavior.
exec "$base/random-game-picker" --play-random >"$PICKER_STATE_DIR/last-run.log" 2>&1
