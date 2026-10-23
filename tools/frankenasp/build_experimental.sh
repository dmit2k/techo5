#!/usr/bin/env bash
# Build the owner-specific AArch64 experimental TECHO5 with a HA-rankable
# version. The release baseline MUST be specified explicitly; no "dev".
set -euo pipefail
if (( $# != 2 )); then
    echo "Usage: $0 vMAJOR.MINOR.PATCH /path/to/output/techo5-experimental" >&2
    exit 2
fi
BASE="$1"
OUT="$2"
[[ "$BASE" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || {
    echo "Expected official base release vX.Y.Z (e.g. v1.6.0)" >&2; exit 2;
}
DIR="$(cd "$(dirname "$0")/../../echod" && pwd)"
command -v go >/dev/null || { echo "Go 1.26 toolchain required" >&2; exit 2; }
command -v git >/dev/null || { echo "git is required" >&2; exit 2; }
COMMIT="$(git -C "$DIR" rev-parse --short=10 HEAD)"
if [[ -n "$(git -C "$DIR" status --porcelain -- internal/hardware/mic ../tools/frankenasp ../echod/internal/update)" ]]; then
    echo "WARNING: experimental source modifications are uncommitted; commit before publishing provenance" >&2
fi
# '_' is stripped for HA's version rank; normal newer upstream releases
# remain visible. A normal latest stable or dev rootfs has no FrankenASP.
VERSION="${BASE}_frankenasp.${COMMIT}"
NOW="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
PKG="github.com/HuskerMinion/techo5/echod/internal/layout"
mkdir -p "$(dirname "$OUT")"
OUT="$(cd "$(dirname "$OUT")" && pwd)/$(basename "$OUT")"
(cd "$DIR" && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath \
    -ldflags "-s -w -X ${PKG}.Version=${VERSION} -X ${PKG}.GitCommit=${COMMIT} -X ${PKG}.BuildDate=${NOW}" \
    -o "$OUT" ./cmd/echod)
echo "Built owner-local experimental TECHO5: $OUT"
echo "Staged software version: $VERSION"
echo "This build is NOT an official upstream TECHO5 release."
