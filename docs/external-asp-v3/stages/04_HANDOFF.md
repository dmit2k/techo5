# 04 — C helper with restored PCM-only factory launcher: coordinator handoff

## 1. Identity / status

- Date: 2026-10-31 (Europe/Riga). Executor: Stage04 C executor on validation host.
- Task: docs/external-asp-v3/stages/04_HELPER_RUNTIME_TASK.md; owner PCM-only Stage04 restart assignment.
- Authoritative docs ref: dmit2k/techo5 experimental/v3-design-staging, df2fc5e9633c1766a4f2921c44d9867e37c9d606.
- Verified SPEC_V2_APPROVED.md blob: f3e1b00d1f5d184e345c0d2c79da8394435ba3a4.
- Correction gate Stage02-R: technical ACCEPTED and integration COMPLETE in published stages/02_PCM_ONLY_CORRECTION_HANDOFF.md; Stage03-R: technical ACCEPTED and integration COMPLETE in published stages/03_PCM_ONLY_CORRECTION_HANDOFF.md.
- Execution result: **PASS (C component software gate, H01–H08)**.
- Stage04 independent delegated technical acceptance: **ACCEPTED** (2026-10-31); C dev-main integration: **COMPLETE**. Executor statuses were PENDING; coordinator independent review and exact remote ref receipt are in §9.
- Hardware: **NOT RUN**; acoustics: **NOT RUN**; real Bionic/vendor Calibrator on Echo: **NOT RUN**. No privileged mount/chroot/device execution performed.
- Stage05: **UNBLOCKED** after accepted/integrated Stage04; separate Stage05 assignment required.

## 2. Exact input / output refs and Git receipt

| Item | TECHO5 Go | FrankenASP V3 C |
|---|---|---|
| Origin | dmit2k/techo5 | dmit2k/frankenasp-v3 |
| Integration ref | experimental/external-asp-v3-dev-main | experimental/v3-dev-main |
| Accepted input integration HEAD | b0c0d5fa96576eff8f7497d48e39fa30a5348e76 | 77149e6494ac9784d98464ff984c84887a28ced9 |
| Stage04 branch | none; unchanged | experimental/v3-stage04 |
| Stage04 worktree | none | ${WORKSPACE}/frankenasp-v3-stage04 |
| Stage04 commits | none | f6266907edb90e7c9c2f414e8e46e6f215a21b7b |
| Tested, published stage candidate | unchanged b0c0d5fa96576eff8f7497d48e39fa30a5348e76 | f6266907edb90e7c9c2f414e8e46e6f215a21b7b |
| Stage04 accepted candidate (delegated technical review) | b0c0d5fa96576eff8f7497d48e39fa30a5348e76 unchanged | **f6266907edb90e7c9c2f414e8e46e6f215a21b7b** |
| Output integration HEAD (coordinator verified) | b0c0d5fa96576eff8f7497d48e39fa30a5348e76 unchanged | **f6266907edb90e7c9c2f414e8e46e6f215a21b7b** (C-only fast-forward) |
| Verified remote main (frozen) | 62603a7854283a3a07473a17c831c6ef0fd6f8d4 | e986ae18643e08e95f29a38aae1dd74762f6362e |

Direct SSH push from validation host: git push --porcelain origin HEAD:refs/heads/experimental/v3-stage04; output new branch (no force), remote GitHub stage ref equals f6266907edb90e7c9c2f414e8e46e6f215a21b7b. GitHub fetch_commit independently returned the same SHA and 27 changed files: https://github.com/dmit2k/frankenasp-v3/commit/f6266907edb90e7c9c2f414e8e46e6f215a21b7b.

Remote Go/C main/dev-main SHAs were rechecked after publication with git ls-remote; same accepted pair and frozen main IDs as above. Historical executor snapshot: no merge, FF, reset, rebase, cherry-pick, force push, code/docs cross-repository write or device action. Subsequently coordinator performed an explicit C-only FF and documented it in §9.

## 3. Contract, exact modules, interfaces and scope

Output pair **T04 = b0c0d5fa96576eff8f7497d48e39fa30a5348e76 (unchanged Go); C04 = f6266907edb90e7c9c2f414e8e46e6f215a21b7b (candidate C)**. This is C component acceptance evidence, not completed Go/C wire integration.

- src/descriptor.[ch]: retained accepted PCM-only 16-byte ASP1 header, 28-byte physical buffer records, 8-byte channel records, maximum 9196 descriptor bytes, one-byte readiness 0xA5; only minimal Bionic include seam added.
- src/runtime.[ch]: helper-side *admission*, strictly confirmed MIC4 S32LE signed24 low/high and REF2 stereo S16LE, 96000 Hz, 1920 input frames per 20 ms, two physical buffers. Generic valid codec configurations without supported vendor geometry reject before READY, including no REF and incompatible temporal geometry. MIC signed24 shift 0/8 and REF L/R conversion occur in C. Trusted captured-slot permutations are explicit owner-provided file values, not deduced from host identity or calibration.
- src/factory_adapter.[ch]: vendor six-function ABI init/create/ref_sync/process/destroy/deinit, 32-bit ARM timespec by value, 768-input-frame to 128-output-sample vendor quantum, guarded buffer bounds, terminal error handling and exactly-once teardown.
- src/vendor_factory.c: production ARM32/Bionic only; dynamically resolves genuine /vendor/lib/libasp.so and six exact exports. No synthetic production backend or mock fallback.
- src/io.[ch]: EINTR, short/full read/write; narrowly scoped owner-authored /factory/channel-map parser. src/child.c: descriptor read, admission, real vendor initialization before READY; private duplicate of stdout wire FD, then ordinary runtime stdout→stderr; immutable PCM full-block exchanges to exactly 320 mono S16LE samples, 640 bytes. No sequence/origin/status/framing.
- src/runtime.c FIFO: one initial 64 zero *output* samples, 768 captured frames→128 DSP output samples per vendor call, 320 mono samples per host response; pending input and output remainder retained between exchanges. No input padding or mute-time reset.
- launcher/launcher.c + launcher/alias_policy.[ch]: native static Linux ARM32 launcher, FD3 trusted executable identity and FD4 owner-owned trusted deploy tree, component-by-component O_NOFOLLOW/owner/mode validation, verified ELF/linker/child/vendor assets, private mount namespace and read-only binds, synthetic proc view with four exact /proc/idme/miccal.N aliases to authentic Device Tree sysfs value files; open/fstat only (no calibration byte access), fail closed on missing necessary resource, fixed environment, CLOEXEC cleanup and execveat of verified Bionic child. This is private factory backend behavior; no generic TECHO5 IDME precondition.
- build_bionic.sh, build_launcher.sh: crossbuild only; no deployment/installer/OTA/device operation. src/bionic_compat.h, src/bionic_entry.c, src/start32.S: genuine freestanding ARM32 Android/Bionic loader seam.
- tests/mock_vendor.c: independently linked test-only ABI mock. tests/io_faults.c: native syscall fault injector. tests/test_runtime.py, tests/test_alias_policy.c, tests/test_launcher.py: focused runtime, alias, security tests. Runtime mock requires test-only ASP1_TEST_MAP and ASP1_TEST_RESOURCE; production binary has neither.
- README.md, LAUNCH_PROVENANCE.md, THIRD_PARTY_NOTICES.md: Stage05 launch prerequisites, precise historical mapping, license. No old bootstrap/control, CFG2/TLV, platform whitelist in TECHO5, status handlers, installers or OTA.
- 27 changed files, 1,415 insertions, 7 deletions vs accepted C input (git diff --stat). Code, tests, and documentation separated in the committed diff. Stock Go/native behavior unchanged; no Go code modified.

## 4. H01–H08 measured test receipts

Evidence root **${WORKSPACE}/frankenasp-v3-stage04-pcmonly-evidence-20261031**. All checked after actual C04 commit on clean worktree with source pair T04/C04 stated above. Commands are from ${WORKSPACE}/frankenasp-v3-stage04 except where stated.

| ID | Executed assertion / command | Exit / measured result | Postcommit evidence SHA256 |
|---|---|---|---|
| H01 | make test; compiled codec probe; positive supported geometry; valid unsupported rate/format/REF absence/temporal before READY; malformed/truncated/legacy descriptor | 0 / PASS | postcommit-H01-H05-H08.log ff887bed27fa08d8896b4fd9c60667a1ce13c9c36fe6b01fc2bb41c411b75c82 |
| H02 | make test: independent REF L/R markers, MIC slot routing and signed24 shift 0/8, mock resource content variation without host/launcher gain; compiled alias policy checks four exact targets, substitution and missing aliases, mock refusal when necessary vendor file is missing | 0 / PASS (offline, not vendor runtime) | same native log |
| H03 | make test: 35 successive full output frames, one 64-sample FIFO seed, 768→128 quantum; 12-block sequence including muted PCM, no reseeding | 0 / PASS | same native log |
| H04 | make test: compiled child descriptor→0xA5→fixed 640 output; 7-byte fragmented writes; native LD_PRELOAD short syscalls and EINTR; startup/PCM EOF, truncated input, broken output pipe; stdout chatter isolated to stderr | 0 / PASS | same native log |
| H05 | make test: independent ABI mock init/create/ref/process and failure scenarios, count/overrun canaries, terminal failures, destroy/deinit exactly once | 0 / PASS | same native log |
| H06 | make asan; make ubsan, independently built codec + full mock C runtime, warnings-as-errors; source/link inventory and production mock-exclusion review | 0 / 0 PASS | postcommit-H06-asan.log c49f039ce755eec16d2e6693353252792dff5ab5205c94c36358be8bd21abe24; postcommit-H06-ubsan.log e14f22abe7f6e1d5a80310d6b78c8cc8208f8b00e71f147d4180194d8013f68d |
| H07 | OUT_DIR=${WORKSPACE}/frankenasp-v3-stage04-pcmonly-evidence-20261031/bionic bash build_bionic.sh, with verified private ARM32 GCC, GNU linker and original Bionic libraries; OUT_DIR=${WORKSPACE}/frankenasp-v3-stage04-pcmonly-evidence-20261031/launcher bash build_launcher.sh with private ARM32 Linux sysroot; readelf ABI/INTERP/NEEDED | 0 / 0 PASS; real production binaries (not mock) | postcommit-H07-bionic.log c381875ea0233d90e24141412a6c92c486fd33f44758ec61d1c8206c443705c9; postcommit-H07-launcher.log dcdf17e782176515b5672668af2a6326e31879260e6f5e2a2f0ddd2ff5131248 |
| H08 | make test launcher source/security suite and compiled alias tests; H08-security-final.log reviews protected FD, private mounts, no old wire/control/installer, production source and strings/ELF dependency inventory; historical license and provenance in Git | 0 / PASS for source/offline boundary | native log above; H08-security-final.log separately included in LOG_SHA256SUMS |

Additional PCM-only fixture receipt: (cd testdata && sha256sum -c SHA256SUMS) exit 0; cmp of C/Go manifests and diff -qr of fixture directories exit 0. Identical original fixture manifest SHA256 **9d26d6cca1e29b10f1abaee583da12ee918ea4252f16f70c3eb2c01531e1d2f4**. Postcommit fixture log SHA256 b2c29c181614052cfa88bd2d79444a70c0f6257fc19c8e4eccac4f9a13f76737; Go/C diff log (empty and exit 0) e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855.

Authoritative postcommit measured log digest index: POSTCOMMIT_LOG_SHA256SUMS SHA256 **129450a43a6365aa05abb560a83744c888a34ba03e705752d002f2f289661b47**. Additional complete precommit/repeated evidence in LOG_SHA256SUMS SHA256 05eecf51c5eb239e32d059841106dc13ed4de2a4d6a5b75500900720e862d2bc. Each includes test identity and actual exit. The earlier trial failures from developing test syntax and compiler search paths are not claimed as successful invocations; the final complete source/postcommit matrix exit 0 is the PASS evidence.

## 5. H07 actual production binary and private toolchain provenance

- Production child: ${WORKSPACE}/frankenasp-v3-stage04-pcmonly-evidence-20261031/bionic/asp1-child, SHA256 **86d4f4d7827129e944c76187dc4575ef7bc5240f2b99f048e7860ed71057fa05**. readelf: ELF32 little-endian ARM EABI5 soft-float, PIE DYN, interpreter /system/bin/linker, DT_NEEDED only libdl.so and libc.so. Vendor libasp.so resolved dynamically at runtime, not vendor bytes in Git; production linked src/vendor_factory.c, NOT tests/mock_vendor.c; binary strings contain no test mode, CFG2 or Q30.
- Production launcher: ${WORKSPACE}/frankenasp-v3-stage04-pcmonly-evidence-20261031/launcher/asp1-launcher-arm32, SHA256 **fce1f3e1abc427d7e5402da5c56dd530d5bf2acd7686b4f3aba98d1445b24ab7**. readelf: native Linux ELF32 ARM EABI5 hard-float EXEC, static; no PT_INTERP, no DT_NEEDED.
- Native codec probe (fixture test): SHA256 423084acb8c414d8178b6ebfab39d3fcc40f61eab549f4fafebb585aa249cf3b; native compiled independent ABI mock: SHA256 5c19a4ee07d25c6821aea43964a3f46ac966310e6ba5e461f97a3e344fdd244d. Production and mock source/link inventories are separate.
- Private compiler ARM32 GCC 13.3.0 at ${WORKSPACE}/techo5-stock-audio-audit-20261021/arm32-tools/cross/usr/bin/arm-linux-gnueabihf-gcc-13, SHA256 73c40a8b836981da0da98dfae66073ed4b9d485cf786899d0c5ef3eeb311e548.
- Private ARM32 GNU ld at ${WORKSPACE}/techo5-stock-audio-audit-20261021/arm32-tools/binutils-unpacked/usr/bin/arm-linux-gnueabihf-ld, SHA256 d15886a5fd7498e8f96a1977ce1b2bbcb2ab617174245d42d026cc2501354bb6.
- Original owner Bionic runtime ${WORKSPACE}/techo5-stock-audio-audit-20261021/factory-arm32-root: libc.so SHA256 e637986bba4c5cc34fb8bf3f143c3b9665bfc6ab509cfd9517844ab87e2d4180; libdl.so SHA256 9807bc310d9d2bd366fc29219b6ce3f1e7e62; ld-android.so SHA256 cfd0e5b8dabe1e30680e3b59c2ab88578e3b2648d9fa96e7d120b143918e9032.
- Linux ARM32 static build sysroot (private, not Git): ${WORKSPACE}/techo5-asp2-f2b-minimal-20261028/private/toolchain/sysroot. Binutils assembler wrapper ${WORKSPACE}/techo5-stock-audio-audit-20261021/arm32-tools/binutils-unpacked/usr/arm-linux-gnueabihf/bin.
- Reproducible exact build variables: BIONIC_CC path above; BIONIC_LD path above; BIONIC_ROOT path above; LD_LIBRARY_PATH=$TC/binutils-unpacked/usr/lib/x86_64-linux-gnu; OUT_DIR=${WORKSPACE}/frankenasp-v3-stage04-pcmonly-evidence-20261031/bionic; bash build_bionic.sh. For launcher: ARM32_CC=BIONIC_CC; ARM32_BINUTILS=$TC/binutils-unpacked/usr/arm-linux-gnueabihf/bin; ARM32_SYSROOT=private Linux path above; OUT_DIR=${WORKSPACE}/frankenasp-v3-stage04-pcmonly-evidence-20261031/launcher; bash build_launcher.sh.
- BUILD_SHA256SUMS (8 binary entries, including sanitizers) SHA256 f98bc5d5d7e636900ce64153cfefaba9b106f15f89f7851dc577a6dbc48052a4.
- H07 is a production **crossbuild verification**, not proof of target firmware execute, vendor resource presence, linker success on an Echo or audible/acoustic correctness.

## 6. Exact source provenance, licensing and negative-scope review

Historical reference only: https://github.com/dmit2k/FrankenASP.git commit **52d51de33e3105501b298c1c214db55f05a2eb82**. Direct historical Git blob IDs, path-by-path reuse map and exact retained launch/ABI fragments are embedded in the published C branch at **LAUNCH_PROVENANCE.md**. Historical applicable MIT license attribution, copyright TECHO5 contributors, full MIT text in **THIRD_PARTY_NOTICES.md**, alongside the new repo's existing MIT LICENSE. Historical references were opened separately, no copy of entire old project, no old commits merged, no proprietary bytes stored.

Allowed transfer: verified component-by-component FD opens and ownership, immutable read-only vendor preparation and native mount namespace/exec, strict alias mapping to genuine device tree without calibration read, Bionic ARM32 entry and six-function ABI/lifecycle, low/high MIC/REF conversion and 768/128 DSP quantum. New V3 codec remains PCM geometry only, and vendor Calibrator is the *sole* factory calibration authority. Launcher, mock and host never serialize or interpret factory Q30, never apply factory PCM gain masks, and no host identity/calibration readers added.

Excluded: CFG2/TLV, 24-byte legacy envelope, HELLO/ACK2/INP2/RSP2, NotReady/status/origin/sequence, installer/OTA/source acquisition, vendor binary packaging, dynamic host platform whitelists, general asset manifest runtime, historical bootstrap, long-lived control plane, speculative resamplers. No device mount or chroot invoked in tests or by scripts.

## 7. Vendor runtime limitations and STOP / conflicts

There is **no code-contract blocker in the tested C Stage04 component**. However, software C PASS is strictly not actual vendor qualification. In particular:
- No actual launcher invocation, private user+mount namespace, real sysfs path resolution, vendor Calibrator calibration read/apply, vendor ABI call on Echo, acoustic acceptance or production Bionic execution occurred.
- Source/security review and offline compiled alias policy test prove expected code and synthetic path behavior, **not** actual root ownership, firmware layout, vendor source-pair authenticity, real runtime linker compatibility or successful device launch.
- Owner must supply and qualify actual root-owned runtime under /data/techo5-linux/external-asp (FD4), deployed static ARM32 launcher (FD3), fixed launcher.profile with OWNER_VERIFIED source pair and supported vendor mode, immutable lib/rootfs including child and Bionic runtime, /factory/channel-map with owner-verified captured-slot permutation, actual vendor configs/libasp.so and four authentic Device Tree miccal.N values. Launcher refuses missing files and exits with stderr before READY; Go should remain Built-in on helper failure once integrated at later stage.
- Without these private, attested resources, **a real factory session is unavailable**. That is a separate vendor-runtime qualification limitation, not H07 build BLOCKED and not permission to claim hardware PASS.
- No Go C runtime paired session or cancellation/deadline/lifecycle integration tested. Host 5-second exchange and 10-second launch deadlines remain the accepted host contract for Stage05/06 to integrate and validate.

## 8. Coordinator publication and handoff gate

The executor report resides outside code/docs worktrees at **${WORKSPACE}/frankenasp-v3-stage04-pcmonly-evidence-20261031/04_HANDOFF.md**. Intended coordinator destination (only after independent review): docs/external-asp-v3/stages/04_HANDOFF.md on dmit2k/techo5 experimental/v3-design-staging. The executor did **not** commit or push docs, advance Go/C dev-main or modify either protected main.

Coordinator to check remote candidate C04 SHA f6266907edb90e7c9c2f414e8e46e6f215a21b7b against C integration 77149e6494ac9784d98464ff984c84887a28ced9 and unchanged Go b0c0d5fa96576eff8f7497d48e39fa30a5348e76, review H01–H08 and provenance, independently decide ACCEPTED/REJECTED, and conduct an explicitly authorized FF-only C integration if accepted. Only after coordinator has published an actual Stage04 ACCEPTED + INTEGRATION COMPLETE receipt may the next separate Stage05 prompt start. Real Echo/deploy/hardware tests remain separately unauthorized.

**Final Stage04 coordinator status: C software PASS; technical acceptance ACCEPTED; C-only integration COMPLETE; hardware/acoustics/vendor device NOT RUN.**

## 9. Independent coordinator acceptance and integration receipt (2026-10-31)

Original unchanged executor handoff: `${WORKSPACE}/frankenasp-v3-stage04-pcmonly-evidence-20261031/04_HANDOFF.md`, SHA256 `824fac2ba5869b8712471eb8209fe223bf16baaeba6d9ea0cd69fa4f996a0a0c`. Historical PENDING refs in the executor's snapshot are superseded by the current verified pair.

**Decision: Stage04 technical ACCEPTED**, strictly for software/C component H01–H08, not real vendor runtime. Reviewed secure factory launcher FD3 executable identity, FD4 root-owned deployment, component O_NOFOLLOW, mount namespace, read-only vendor assets and four private `/proc/idme/miccal.N` aliases without reading factory bytes; C child uses genuine six-function vendor ABI and PCM-only descriptor, real vendor initialization before READY, independent REF L/R, signed24 alignment, 768→128 DSP quanta, one-time 64 output samples, 640-byte fixed PCM. Factory-only admission does not add host platform restrictions.

Independent C replay: `make test`, `make asan`, `make ubsan` **PASS**. Postcommit logs and eight binary SHA256 entries verified; seven historical provenance blobs compared to actual Git objects at `52d51de33e3105501b298c1c214db55f05a2eb82`; MIT notices present. Tracked `testdata/*.bin` are synthetic PCM fixtures, not private factory data. No vendor lib, calibration, or deploy assets in Git.

**Independent H07 crossbuild:** coordinator rebuilt production Bionic child and native ARM32 static launcher using existing private ARM32 toolchain/Bionic libraries/Linux sysroot. Both outputs **byte-identical** to executor's originals: child SHA256 `86d4f4d7827129e944c76187dc4575ef7bc5240f2b99f048e7860ed71057fa05` and launcher SHA256 `fce1f3e1abc427d7e5402da5c56dd530d5bf2acd7686b4f3aba98d1445b24ab7`. Artifacts: `.../frankenasp-v3-stage04-pcmonly-evidence-20261031/coordinator-review-builds/`. No real launcher mount/chroot, vendor libasp.so, Echo, Calibrator or acoustic test performed or claimed.

**C-only fast-forward integration:** clean preflight refs Go dev-main `b0c0d5fa96576eff8f7497d48e39fa30a5348e76`, C dev-main `77149e6494ac9784d98464ff984c84887a28ced9`, C stage ref `f6266907edb90e7c9c2f414e8e46e6f215a21b7b`; one parent commit. Coordinator executed `git merge --ff-only f6266907edb90e7c9c2f414e8e46e6f215a21b7b` in `${WORKSPACE}/frankenasp-v3-dev-main`, direct `git push --porcelain origin HEAD:refs/heads/experimental/v3-dev-main` from `validation host`. Remote C dev-main and C stage ref both confirmed C04, Go dev-main unchanged. Protected main HEADs frozen at TECHO5 `62603a7854283a3a07473a17c831c6ef0fd6f8d4` and FrankenASP `e986ae18643e08e95f29a38aae1dd74762f6362e`. No force/reset/rebase/device changes; worktrees clean.

**Accepted Stage05 input: Go `b0c0d5fa96576eff8f7497d48e39fa30a5348e76`, C `f6266907edb90e7c9c2f414e8e46e6f215a21b7b`.** Technical acceptance **ACCEPTED**, integration **COMPLETE**. Stage05 may begin its own paired Go/C wire/session implementation. Real vendor/target qualification remains explicitly not done.
