#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_dir=$(dirname "$script_dir")
cd "$repo_dir"

version=${VERSION:-}
commit=${COMMIT:-unknown}
build_date=${BUILD_DATE:-unknown}
dist_dir=${DIST_DIR:-dist}
targets=${TARGETS:-"linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64"}
dry_run=false
temporary_dir=""

cleanup() {
  if [ -n "$temporary_dir" ]; then
    rm -rf "$temporary_dir"
  fi
}
trap cleanup EXIT HUP INT TERM

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

ldflags="-s -w -X github.com/cinnamorollofficials/go-code-scanner/pkg/buildinfo.Version=$version -X github.com/cinnamorollofficials/go-code-scanner/pkg/buildinfo.Commit=$commit -X github.com/cinnamorollofficials/go-code-scanner/pkg/buildinfo.Date=$build_date"

if [ "$dry_run" = false ]; then
  case "$build_date" in
    ????-??-??T??:??:??Z) ;;
    *) echo "BUILD_DATE must be an RFC3339 UTC timestamp" >&2; exit 2 ;;
  esac
  mkdir -p "$dist_dir"
  temporary_dir=$(mktemp -d "${TMPDIR:-/tmp}/go-code-scanner-release.XXXXXX")
fi

for target in $targets; do
  goos=${target%/*}
  goarch=${target#*/}
  suffix=""
  if [ "$goos" = "windows" ]; then
    suffix=".exe"
  fi
  name="security-review_${version}_${goos}_${goarch}${suffix}"
  extension=".tar.gz"
  if [ "$goos" = "windows" ]; then
    extension=".zip"
  fi
  archive="$dist_dir/security-review_${version}_${goos}_${goarch}${extension}"
  if [ "$dry_run" = true ]; then
    echo "GOOS=$goos GOARCH=$goarch go build -trimpath -buildvcs=false -ldflags=<release-metadata> -o <temporary>/$name ./cmd/security-review && archive $archive"
    continue
  fi
  output="$temporary_dir/$name"
  GOOS=$goos GOARCH=$goarch CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags "$ldflags" -o "$output" ./cmd/security-review
  go run ./scripts/archive_release.go "$output" "$archive" "$build_date"
done
