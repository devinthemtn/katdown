#!/usr/bin/env bash
# Runs the Windows katdown.exe from WSL, turning file arguments into Windows
# paths. wails3 task install-windows fills in the exe path.
exe='@EXE@'
args=()
foreground=
for a in "$@"; do
  if [[ -e "$a" ]]; then
    args+=("$(wslpath -w "$(realpath "$a")")")
  else
    args+=("$a")
    # Any flag but -html (-version, -makeconfig, -h, a typo, ...) prints
    # something and exits instead of opening a window.
    case "$a" in
      -html | --html | -html=* | --html=*) ;;
      -*) foreground=1 ;;
    esac
  fi
done
if [[ -n $foreground ]]; then
  exec "$exe" "${args[@]}"
fi
# Detached, so the shell isn't held while the window is open.
nohup "$exe" "${args[@]}" >/dev/null 2>&1 &
disown
