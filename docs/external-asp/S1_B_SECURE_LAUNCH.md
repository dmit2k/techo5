# S1-B secure discovery and host launch (offline only)

S1-B composes the S1-A1 inode resolver and S1-A2 FD execution with the
existing supervised Go host. It does not qualify or activate a factory helper.
The only production transport remains the legacy 40-byte configuration.

## Selection and cache

`discoveredASP` is a once-per-daemon typed selection. A secure candidate carries
executable and asset-root dev/inode pairs, never cached open FDs or a string
that can become a legacy candidate. Discovery closes its temporary pins.

Only `openSecureExternalASP` can authorize exactly
`/data/techo5-linux/external-asp/bin/external-asp`. It walks from `/` with the
unchanged S1-A1 UID/mode/openat policy. Legacy `findASP` explicitly excludes
that path and symlinks resolving to it. The old `/data/techo5-linux/bin` path,
PATH and sibling candidates retain historical strict ancestor checks, including
the existing root-owned sticky temporary-directory rule. No 0777 ancestor
exception is added to any legacy candidate.

The production discovery call supplies `qualified=false`. A valid secure-only
candidate is recorded as unqualified: `HaveExternalASP` returns false and
activation fails closed without reopening or executing it. An independently
strict legacy candidate can still be advertised on Athena. An invalid secure
candidate is never evidence that a legacy candidate succeeded. `HaveExternalASP`
is cached advertised capability, not a promise that later startup cannot fail.

Offline tests alone call the private selector with `qualified=true`. There is
no ENV, settings or HELLO route to this opt-in. Qualification and transport are
independent: a qualified synthetic choice still defaults to legacy 40 bytes.
CFG2/TLV requires an explicitly selected transport and a fully validated,
immutable session matching MIC/REF and host metadata rules before launch.

Every secure activation repeats the secure open and both inode comparisons.
A changed name before the activation pin fails closed, even if the replacement
also passes ownership checks. A changed name after the pin executes the opened
ELF with the opened asset root. There is no pathname retry, format downgrade,
rediscovery or automatic restart after failure. Replacing an installation
requires an explicit daemon restart and later qualification policy.

## Process ownership

`prepareSecureASPPinned` creates an unstarted command using literal
`/proc/self/fd/3`, `ExtraFiles=[executable, externalRoot]`, and only LANG/LC_ALL.
The host creates and attaches its actual stdin/stdout os.Pipe descriptors before
its single `Cmd.Start`. Launch-only parent descriptors close immediately after
Start on both success and failure, and on earlier context/pipe errors. The
child owns independent FD3/FD4 copies. No global environment is changed.

The same host reaper calls Wait once. Cancellation/Close closes the host pipe
ends and kills the direct child; capture never waits on reaping. HELLO remains
500 ms. CFG/ACK/first exchange share one absolute 25 ms deadline; each later
exchange has 25 ms. Source's single startup worker, epoch rules, three DATA
preflight frames, timeline validation and current-tick builtin fallback are
unchanged. Source serializes launch/close; the pinned object remains exclusively
owned by its startup caller until Start returns.

## Evidence and limits

The new independent C fixture implements actual HELLO/CFG/ACK/INP/RSP without
using Go codecs. It checks FD3 against `/proc/self/exe`, reads the profile through
component-wise `openat` from FD4, and rejects inherited environment beyond the
minimal contract. Its response encodes a CRC of every raw MIC/REF byte plus
separate compiled-code and asset markers, distinguishing original and decoy.
It checks both exact legacy bytes and explicitly selected CFG2 descriptors.
Existing golden vectors and the PCM/metadata protocol implementations are
unchanged. No real IDME reader or vendor DSP is used.

FD pinning is not executable authenticity, a process sandbox or a loader
sandbox. PT_INTERP and shared libraries remain system dependencies. The host
provides FD4; the real launcher must independently enforce asset traversal and
runtime isolation in a later stage. A privileged writer mutating trusted inode
contents remains outside the S1-A1 threat model. Legacy pathname execution
retains its historical check-to-exec limitation; writable-parent exceptions
remain exclusive to the secure resolver.

Missing procfd is tested through the existing deliberately nonexistent
`/proc/self/fd/-1` seam; no mount/chroot/procfs changes are made. A synthetic
bounded descendant holds inherited pipes after direct-child exit: HELLO still
times out, and the direct child is reaped. Arbitrary descendants are not killed
or reaped by this host; closing our descriptors bounds host I/O regardless.
Native x86_64 ELF evidence and Linux ARMv7 cross-builds do not qualify ARM32
kernel/procfs/Bionic execution or future factory assets. F1 is not part of S1-B.
