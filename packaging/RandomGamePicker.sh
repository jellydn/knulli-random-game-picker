#!/bin/sh
set -eu

base=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
export PICKER_STATE_DIR="$base/state"
mkdir -p "$PICKER_STATE_DIR"
umask 077
# Replace the previous session log instead of appending across sessions.
exec "$base/random-game-picker" >"$PICKER_STATE_DIR/last-run.log" 2>&1
