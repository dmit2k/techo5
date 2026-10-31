# Android NDK r28c — installed permanently on anfisa-trainer

**Status:** INSTALLED / VERIFIED 2026-10-31, gptlab user, no sudo and no system-wide package changes. This is not a future installation task.

| Item | Persistent path or version |
|---|---|
| NDK | `/home/gptlab/Android/Sdk/ndk/28.2.13676358` |
| `ANDROID_SDK_ROOT` / `ANDROID_HOME` | `/home/gptlab/Android/Sdk` |
| `ANDROID_NDK_HOME` / `ANDROID_NDK_ROOT` | `/home/gptlab/Android/Sdk/ndk/28.2.13676358` |
| Environment to source | `/home/gptlab/.config/techo5/toolchains/android-ndk.env.sh` |
| Startup | `gptlab` `~/.profile` + `~/.bashrc` source this environment |
| Agent-wide discoverability | `/home/gptlab/dev/ANDROID_NDK_TOOLCHAIN.md` and `~/.config/techo5/toolchains/README.md` |
| Original cached ZIP | `/home/gptlab/.cache/techo5-toolchains/android-ndk-r28c-linux.zip` |
| Official source | https://dl.google.com/android/repository/android-ndk-r28c-linux.zip |
| Google SHA1 verified | `a7b54a5de87fecd125a17d54f73c446199e72a64` |
| Local ZIP SHA256 | `dfb20d396df28ca02a8c708314b814a4d961dc9074f9a161932746f815aa552f` |
| Full test receipt | `/home/gptlab/dev/techo5-v3-stage07-ndk-validation-20261031/NDK_INSTALL_RECEIPT.md` |

For agents and SSH shells, **explicitly source** the environment before testing:

```sh
. /home/gptlab/.config/techo5/toolchains/android-ndk.env.sh
"$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android26-clang" --version
"$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/linux-x86_64/bin/armv7a-linux-androideabi26-clang" --version
```

Verified `source.properties` revision **28.2.13676358**, Clang **19.0.1**, target API26, both successful actual Android linker smoke checks (ELF64 /system/bin/linker64, ELF32 /system/bin/linker). On the pre-correction Go Stage06 SHA `cce2f220aeab9835e10a7f7ab659f5ca390011e8`, **two real `GOOS=android CGO_ENABLED=1 -tags panel` TECHO5 builds PASS**, using Go 1.26.9:
- arm64-v8a ELF64 SHA256 `ea42091b7b1c342d3a1720a5f0237b29d129166e4b0c12494f9dea0d156a5937`
- armeabi-v7a ELF32 SHA256 `3e211441de2b844a5faa53259305129835ae8693ce452b2a132295e699c64bd3`

Real binaries and SHA256SUMS: `/home/gptlab/dev/techo5-v3-stage07-ndk-validation-20261031/`. Standalone NDK includes compilers and sysroots. SDK manager/platform packages are **not installed**; they were not needed for the two direct Go/Cgo Panel crossbuilds.

**Acceptance separation:** the old Stage07 S08 missing-NDK *environmental blocker* has been removed. **Stage07 software verdict remains FAIL/STOP** because S07 dot test geometry still requires correction, independent acceptance/integration, and full retake S01–S10 on the new exact pair. Neither these preliminary ABI outputs nor compiler installation proves retake PASS, live Echo, vendor DSP, or acoustics.
