# Portable verification environment

The published engineering notes use symbolic paths rather than machine-specific directories.

- `WORKSPACE`: local workspace holding independently checked-out Go/C sources and test evidence. It is not a repository constant.
- `ANDROID_NDK_HOME`, `ANDROID_SDK_ROOT`: locally provisioned Android SDK/NDK roots; discover and verify `source.properties` for the required toolchain version.
- `TOOLCHAIN_ENV_FILE`: optional local shell file used to export toolchain paths. The file is configured outside Git.
- `USER_CACHE_DIR`, `USER_CONFIG_HOME`: local cache and configuration locations that must not be included in published artifacts.

Example for a worker:

```sh
: "${WORKSPACE:?configure local workspace}"
: "${ANDROID_NDK_HOME:?configure validated Android NDK}"
test -f "${ANDROID_NDK_HOME}/source.properties"
```

Do not commit local home directories, internal hostnames, SSH endpoints, evidence storage roots, machine inventories, or site-specific execution receipts. This document is portable: exact local paths belong only in an access-controlled runner's environment.
