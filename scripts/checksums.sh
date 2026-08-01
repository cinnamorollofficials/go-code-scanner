#!/bin/sh
set -eu

dist_dir=${DIST_DIR:-dist}
manifest="$dist_dir/SHA256SUMS"
temporary="$dist_dir/.SHA256SUMS.tmp"

if [ ! -d "$dist_dir" ]; then
  echo "release directory does not exist: $dist_dir" >&2
  exit 2
fi

cleanup() {
  rm -f -- "$temporary"
}
trap cleanup EXIT INT TERM
: > "$temporary"

for artifact in "$dist_dir"/*; do
  [ -f "$artifact" ] || continue
  [ "$artifact" = "$manifest" ] && continue
  name=${artifact##*/}
  if command -v sha256sum >/dev/null 2>&1; then
    digest=$(sha256sum "$artifact" | awk '{print $1}')
  else
    digest=$(shasum -a 256 "$artifact" | awk '{print $1}')
  fi
  printf '%s  %s\n' "$digest" "$name" >> "$temporary"
done

LC_ALL=C sort -k2,2 -o "$temporary" "$temporary"
mv -f -- "$temporary" "$manifest"
trap - EXIT INT TERM
echo "wrote $manifest"
