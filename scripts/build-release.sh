#!/bin/sh
set -eu

version=${VERSION:-}
commit=${COMMIT:-unknown}
build_date=${BUILD_DATE:-unknown}
dist_dir=${DIST_DIR:-dist}
targets=${TARGETS:-"linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64"}
dry_run=false

if [ "${1:-}" = "--dry-run" ]; then
  dry_run=true
elif [ -n "${1:-}" ]; then
  echo "usage: VERSION=v1.2.3 scripts/build-release.sh [--dry-run]" >&2
  exit 2
fi

case "$version" in
  v[0-9]*.[0-9]*.[0-9]*) ;;
  *) echo "VERSION must be a semantic version such as v1.2.3" >&2; exit 2 ;;
esac

ldflags="-s -w -X github.com/cinnamorollofficials/go-code-scanner/buildinfo.Version=$version -X github.com/cinnamorollofficials/go-code-scanner/buildinfo.Commit=$commit -X github.com/cinnamorollofficials/go-code-scanner/buildinfo.Date=$build_date"

for target in $targets; do
  goos=${target%/*}
  goarch=${target#*/}
  suffix=""
  if [ "$goos" = "windows" ]; then
    suffix=".exe"
  fi
  output="$dist_dir/security-review_${version}_${goos}_${goarch}${suffix}"
  if [ "$dry_run" = true ]; then
    echo "GOOS=$goos GOARCH=$goarch go build -trimpath -buildvcs=false -ldflags=<release-metadata> -o $output ./cmd/security-review"
    continue
  fi
  mkdir -p "$dist_dir"
  GOOS=$goos GOARCH=$goarch CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags "$ldflags" -o "$output" ./cmd/security-review
done
