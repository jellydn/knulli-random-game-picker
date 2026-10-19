#!/bin/sh
set -eu

base=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
export PICKER_STATE_DIR="$base/state"
export PICKER_CONTROLLER_LAYOUT=knulli
# The Ports hook and configgen use this flag for the non-swapped SDL mapping.
if [ -e /var/run/ports_xbox_layout.flag ]; then
    export PICKER_CONTROLLER_LAYOUT=xbox
fi
mkdir -p "$PICKER_STATE_DIR"
umask 077
# Replace the previous session log instead of appending across sessions.
exec "$base/random-game-picker" >"$PICKER_STATE_DIR/last-run.log" 2>&1
