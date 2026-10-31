#!/usr/bin/env bash
# Runs the Windows katdown.exe from WSL, turning file arguments into Windows
# paths. wails3 task install-windows fills in the exe path.
exe='@EXE@'
args=()
for a in "$@"; do
  if [[ -e "$a" ]]; then
    args+=("$(wslpath -w "$(realpath "$a")")")
  else
    args+=("$a")
  fi
done
# Detached, so the shell isn't held while the window is open.
nohup "$exe" "${args[@]}" >/dev/null 2>&1 &
disown
