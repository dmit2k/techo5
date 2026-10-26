# External ASP protocol and channel-mapping contract

> **Experimental ASP2, not a frozen public ABI.** This document reflects the offline-tested generic P3-3A transport and synthetic adapter. A P3-3A.1 revision changes the *host selection policy* but is intended to preserve these wire formats. No genuine factory ASP runtime or microphone calibration is bundled.

## 1. Distinguish ASP1 from ASP2

**ASP1 is a frozen P3-2 comparison fixture**, not the generic multichannel format. It sends exactly 4 mic channels and 2 stereo hardware-reference channels at 96 kHz for 20 ms, then expects 320 mono samples immediately. It has no negotiable input topology, status envelope, declared output origin or `NOT_READY`.

**ASP2 is the generic host/helper boundary.** A once-per-process `CFG2` descriptor declares independently counted MIC and REF streams; `ACK2` accepts/rejects a particular immutable stream shape; `INP2` sends the captured slots without vendor-specific reordering; `RSP2` returns 16-kHz mono audio with bounded status and an explicit input-time origin.

There is no automatic ASP2→ASP1 downgrade. A helper implementing only ASP1 is not ASP2-compatible.

| Property | ASP1 | ASP2 |
|---|---|---|
| Input topology | Fixed 4 MIC + 2 REF | 1–16 MIC; 0–8 REF, subject to payload cap |
| MIC/REF rates, types, scaling | Fixed contract | Negotiated per stream |
| Captured slot order | Interleaved, fixed native capture order | Interleaved, immutable capture order |
| Geometry / mic mapping | Outside wire | Still outside wire: helper-local mapping |
| Initial DSP not-ready | Not representable | Explicit `NOT_READY` |
| PCM input-time origin | Not present | `pcmStartSample` |
| Structured status / rejection | Not present | `ACK2` and `RSP2` statuses |
| Production stability | Prototype comparison | Experimental review contract |

## 2. Input channel semantics: what each number means

Never equate these concepts:

- **Physical sensors:** the microphones mounted in a product, not necessarily individual independent output slots.
- **Captured MIC slots:** channels provided by native capture, numbered from 0 to `micChannels-1` and preserved in their *actual captured* order.
- **Captured REF slots:** hardware playback-reference channels, independently numbered from 0 to `refChannels-1`. A missing reference means **zero REF channels**, not two zero-filled fabricated channels.
- **DSP logical inputs:** ordered positions required by one helper's beam geometry, vendor algorithm and calibration. Their count/order can differ from captured input.

The host declares **captured** shape, not geometry. The helper must verify whether captured slots can be mapped to its DSP requirements. Capture slot indices remain invariant for the lifetime of an acknowledged helper session.

### Athena's current native capture tap

The present TECHO5 source-owned Athena tap advertises:

| Stream | Channel count | Rate | Wire format | Meaning |
|---|---:|---:|---|---|
| MIC | 4 | 96,000 Hz | S32_LE, 1,920 frames / 20 ms | Capture slot 0, 1, 2, 3 in ADC-provided order |
| REF | 2 | 96,000 Hz | S16_LE, 1,920 frames / 20 ms | Capture playback-reference slot 0 (nominal L), slot 1 (nominal R) |

Both are interleaved *within their own stream*, not six channels in a single input array. `INP2` contains the complete interleaved MIC plane **followed by** the complete interleaved REF plane:

```text
MIC: [sample0/mic0, sample0/mic1, sample0/mic2, sample0/mic3,
      sample1/mic0, sample1/mic1, sample1/mic2, sample1/mic3, ...]
REF: [sample0/ref0, sample0/ref1, sample1/ref0, sample1/ref1, ...]
Wire INP2 payload = MIC bytes || REF bytes
```

Do **not** infer that ADC capture slot 0 corresponds to a particular physical direction on the enclosure: the physical microphone-to-slot map has not yet been qualified. Similarly, ref0/ref1 are the ALSA stereo slot order; any hardware-side L/R verification or swap required by a specific DSP is an adapter integration responsibility.

The **built-in** 16-kHz converted reference has a different layout: its first reference position contains the mixed `(L+R)/2` signal and its second contains R. That built-in representation must **never** be fed to an external helper while pretending to be the native independent stereo hardware reference.

For Athena on the factory kernel, the MIC 32-bit container holds **low-aligned 24-bit sign-extended samples**, valid bits 24, exponent 23; on the tested alternate kernel configuration, samples are **high-aligned**, valid bits 24, exponent 31. Both MIC paths describe the *existing signed samples*, not a newly boosted, downsampled or recalibrated signal. REF uses signed S16 with 16 valid bits, exponent 15. The helper must not apply an unconditional `>>8` to both MIC cases.

## 3. ASP2 session and byte encoding

All multibyte integers are little-endian. Stdout is protocol-only; stderr carries diagnostic output. A helper is a trusted local executable, owned and supervised by the TECHO5 process.

Session:

1. Helper writes four ASCII bytes `ASP2` (HELLO) within the existing **500-ms** asynchronous startup budget.
2. When the **first actual native capture frame** exists, host writes exactly one `CFG2` descriptor and reads `ACK2` ConfigResult.
3. After ACCEPT, host sends a sequenced `INP2` request and reads its corresponding `RSP2`.
4. No stream descriptor, mapping or geometry changes are permitted mid-session. A new process starts a new session.

`CFG2` negotiation and the first PCM exchange share **one overall 25-ms** provisional capture-path deadline. Subsequent request+reply exchanges each share the same budget. There is no separate long-running configuration phase inside the capture tick.

### 3.1 CFG2 (40 bytes, once per helper instance)

| Offset | Type | Field |
|---:|---|---|
| 0 | char[4] | ASCII `CFG2` |
| 4 | uint16 | Protocol version = 2 |
| 6 | uint16 | Descriptor bytes = 40 |
| 8 | uint32 | Flags |
| 12 | uint16 | MIC channel count |
| 14 | uint16 | REF channel count |
| 16 | uint32 | MIC sample rate (Hz) |
| 20 | uint32 | REF sample rate (Hz) |
| 24 | uint16 | MIC frames per 20-ms tick |
| 26 | uint16 | REF frames per 20-ms tick |
| 28 | uint8 | MIC sample format |
| 29 | uint8 | REF sample format |
| 30 | uint8 | MIC full-scale exponent |
| 31 | uint8 | REF full-scale exponent |
| 32 | uint8 | MIC alignment |
| 33 | uint8 | REF alignment |
| 34 | uint8 | MIC valid bits including sign |
| 35 | uint8 | REF valid bits including sign |
| 36 | uint32 | Reserved, must be zero |

Flags: `INTERLEAVED=1`, `SYNCHRONIZED=2`, `HARDWARE_REF=4`.

With REF channels present, flags **must equal 7**: this indicates a real, independently acquired and synchronized hardware reference. With REF count 0, flags **must equal 1** and all REF descriptor fields must be zero. No synthetic zero-filled or software-ref channel may be falsely advertised as hardware reference.

Per-stream format `0=NONE` (only the absent REF descriptor), `1=S16_LE`, `2=S32_LE`. Alignment `0=LOW`, `1=HIGH`. For LOW, full-scale exponent equals `validBits-1`; for HIGH it equals `containerBits-1`.

Limits checked by the host:

- MIC channel count 1–16; REF 0–8.
- Each populated stream rate 8,000–192,000 Hz and divisible by 50; per-tick frames must equal `rate/50`.
- Valid bits in the range 2 to container bit width; signed sample format and scale must agree.
- MIC+REF data payload ≤262,144 bytes (256 KiB) per 20-ms tick, computed with overflow-safe arithmetic.
- Both streams' channel counts and formats remain unchanged for the session.

These are **wire limits**, not proof of ALSA support for 16 mics or of any vendor DSP handling them. Offline tests exercise MIC counts 1–8 × REF 0–2 and additional boundaries; current physical Athena native capture still supplies 4+2.

### 3.2 ACK2 / ConfigResult (12-byte header + bounded message)

| Offset | Type | Field |
|---:|---|---|
| 0 | char[4] | ASCII `ACK2` |
| 4 | uint8 | ACCEPT=1 or REJECT=2 |
| 5 | uint8 | Status kind: NONE=0, STATE=1, EVENT=2 |
| 6 | uint16 | Optional status code |
| 8 | uint16 | Status text bytes, 0–160 |
| 10 | uint16 | Capability bits: NO_REFERENCE=1; others zero |
| 12 | bytes | UTF-8 status text |

For a 0-REF stream, ACCEPT must explicitly advertise `NO_REFERENCE` support. For nonzero REF, that flag must be zero.

ACCEPT signals a validated topology and capability, **not** a valid PCM reply and not readiness. A fatal `5xx` status is terminal regardless of accompanying ACCEPT/REJECT. A REJECT without a code is terminal by type. Recoverable `3xx/4xx` messages can accompany accepted operation only if the helper truly has a verified safe degraded mode; they are not permission to invent missing mics or reference channels.

### 3.3 INP2 (20-byte header + native data)

| Offset | Type | Field |
|---:|---|---|
| 0 | char[4] | ASCII `INP2` |
| 4 | uint64 | Request sequence, starting at zero |
| 12 | uint32 | MIC payload length, equals descriptor |
| 16 | uint32 | REF payload length, equals descriptor |
| 20 | bytes | MIC samples followed by REF samples |

The MIC and REF planes are separately interleaved in the original native captured slot order. There are no per-tick channel maps, board identifiers, calibration coefficients or geometry fields. REF payload is truly absent if the descriptor declares zero channels.

The request sequence identifies IPC order. It is not a timestamp, a device epoch or proof of acoustic timing.

### 3.4 RSP2 (32-byte header + bounded message + optional PCM)

| Offset | Type | Field |
|---:|---|---|
| 0 | char[4] | ASCII `RSP2` |
| 4 | uint8 | DATA=1, NOT_READY=2, FATAL=3 |
| 5 | uint8 | Status kind |
| 6 | uint16 | Status code |
| 8 | uint16 | Status text bytes, 0–160 |
| 10 | uint16 | Reserved, must be zero |
| 12 | uint64 | Echoed request sequence |
| 20 | uint64 | `pcmStartSample` |
| 28 | uint32 | PCM payload length |
| 32 | bytes | Status text, followed by optional PCM16 |

DATA contains exactly **640 bytes = 320 S16_LE mono samples at 16 kHz**. Sample-origin positions are in the 16-kHz **input-time** coordinate system, not the internal algorithmic latency of the vendor DSP.

`NOT_READY` has no PCM, length zero, `pcmStartSample=UINT64_MAX`; in the currently implemented seam it is permitted **only on the first request**. The next and following responses must be valid DATA. Initial DATA must start at origin zero; subsequent DATA origins advance by exactly 320. Delayed data may lag input by at most one 20-ms tick in this supported design. A larger arbitrary buffering latency is not silently allowed.

A valid fully framed fatal 5xx status is acted upon **without waiting for appended PCM**: the host terminates the helper and selects stock. A malformed status, invalid UTF-8, bad message length, wrong sequence/origin or unexpected DATA/NOT_READY combination also causes fallback.

## 4. Mapping captured channels to logical DSP inputs

**There are no channel-map arrays in ASP2.** Source-slot mapping is part of the helper's local configuration and DSP profile, not TECHO5's HA selection or persisted microphone settings.

Illustrative helper-owned configuration (no default file location standardized):

```json
{
  "micChannelMap": [2, 0, 3, 1],
  "refChannelMap": [1, 0]
}
```

Interpret each array index as a **DSP destination** and each value as a **captured source slot**:

| DSP logical input | Captured slot supplied |
|---|---|
| mic0 | mic2 |
| mic1 | mic0 |
| mic2 | mic3 |
| mic3 | mic1 |
| ref0 | ref1 |
| ref1 | ref0 |

Requirements for a given helper profile:

- The MIC map length must equal the DSP's required number of logical MIC inputs; the REF map length similarly matches logical REF.
- Every captured source index must exist, be nonnegative and occur **at most once** within that map.
- REF and MIC maps are distinct; a reference slot cannot substitute for a mic or vice versa.
- A verified profile may explicitly permit selecting a subset (for example, capture 8 mic slots → a 4-input DSP using `[7,3,0,6]`); implicit truncation, duplicated slots and fabricated inputs are forbidden.
- Omitting a map implies identity **only when source order, logical geometry and calibration have been independently verified** for that exact profile.
- The mapping and DSP profile are immutable once the stream descriptor is acknowledged.
- Rearrangement changes slot selection only; it does not resample, shift timestamps, flip sample sign, reverse time or add hidden gain.

For a factory-shaped 4+2 DSP, the synthetic adapter prepares destination samples in the order:

```text
DSP input for native time t:
  logical_mic0, logical_mic1, logical_mic2, logical_mic3,
  logical_ref0, logical_ref1
```

This **six-slot interleave belongs to the helper**. The host's `INP2` payload remains separate MIC and REF planes in capture order.

The experimental `internal/asp/synthetic` code implements profile validation, permutations, subset selection, and captured sample scaling in **offline fixtures**. The real factory-specific channel map, physical sensor geometry, internal calibration coupling, helper configuration file policy and native vendor runtime are **not** established by those tests. The factory-shaped synthetic profile accepts only 4 MIC + 2 real 96-kHz REF; generic test profiles can accept more.

### Calibration, scale and channel identity

Mic calibration may index physical capture slots or vendor logical DSP inputs. When applying a permutation, a gain or phase correction must remain associated with its intended physical signal; do not apply miccal twice or reorder calibration blindly. Actual vendor calibration ownership and accessibility must be verified during device integration.

The synthetic factory-shaped conversion maps full-scale-23 low-aligned microphone samples to the appropriate factory 24-bit scale, or handles full-scale-31 high-aligned values with an explicit shift, with clipping where specified. Native REF scaling is separately handled. This does not license adopting a single hardcoded shift for all boards.

A helper missing verified topology/geometry/calibration must REJECT with a terminal `5xx` diagnostic rather than guessing an identity order.

## 5. Causal 8-ms DSP adaptation

A DSP that consumes **768 input samples/channel at 96 kHz** and produces **128 output samples at 16 kHz** cannot synchronously produce the full 320 output samples from its first 1,920-input-sample TECHO5 tick.

The synthetic FIFO consumes native input exactly once using two, then three, then two, then three 8-ms quanta across consecutive 20-ms ticks. It returns an initial NOT_READY and then sequential 320-sample DATA blocks with explicit origins. The residual native and PCM buffers are bounded; no fake initial 64 samples, no zero padding and no arbitrary frame cloning are needed to make output sizes match.

The **selection policy at the transition** is separate from this DSP FIFO. The earlier all-samples-continuous policy inserted a publication pause and delayed stock shadow; the low-latency-first revision instead prioritizes one current-tick stock publication on every tick and accepts only bounded origin overlap/skip **between different processors**. Within a stable External run, validated DSP-origin frames must be continuous with no duplicated/skipped sample positions.

Hardware processor group delay, speaker-to-microphone acoustic delay and ALSA mic/ref start offset must be measured separately. The transport origin alone is not an end-to-end latency measurement.

## 6. Status semantics

Status messages are structured into NONE, STATE and EVENT kinds with numeric code and optional bounded plain text.

- `000`: host/default context, not proof of DSP recovery.
- `2xx`: informational/healthy state.
- `3xx` and `4xx`: recoverable warnings, only with a genuinely supported continuing mode.
- `5xx`: terminal failure regardless of where the status is carried.

User-facing diagnostic form is exactly `XXX: MESSAGE_TEXT` (with a three-digit code). State changes are de-duplicated; EVENT messages are subject to throttling for Home Assistant display. Valid nonfatal warnings remain visible until a meaningful healthy state or a new attempt; numeric codes and free-form text are not persisted in the configuration Store.

Status UTF-8 text is limited to 160 bytes, and unsafe formatting/control/path-like content is conservatively rejected or replaced with a safe default. The helper must not emit secrets, local paths, IPs, personal data or proprietary payloads through status.

The helper's stderr is for ordinary diagnostics; its stdout must never contain `printf` output unrelated to framed ASP2.

## 7. Mandatory conformance tests before real DSP use

Test topology and routing before acoustic performance:

1. MIC 1–8 × REF 0/1/2 (24 shapes), plus boundary 9/12/16 MIC and REF 8; enforce the total payload limit.
2. Nontrivial permutations for 4, 7 and 8 MIC, both REF swap and true no-reference operation; correctly reject unsupported shapes.
3. Malformed maps, missing geometry, unsafe default identity, duplicate/out-of-range indices and runtime mutation.
4. Signed sample marker/impulse tests for source slot preservation, logical destination assignment, valid-bit scaling and calibration ownership.
5. No zero-filled hardware-reference fabrication, no mixing REF into MIC, and explicit no-reference capability ACK.
6. Exact framing, partial I/O, sequence, payload lengths, origins, NOT_READY, fatal 5xx and process reaping.
7. Three live-DATA preflight frames, last-command-wins, HA requested/active truthfulness and stock fallback on the failing tick.
8. 8→20 FIFO sample accounting over long runs, no unbounded queue and measurable publication cadence.
9. Bit-for-bit stock equivalence when External is absent or inactive, unaffected WebRTC path, and portability tests across build targets.
10. Real hardware calibration, capture clock/reference alignment, scheduler behavior, CPU cost, software/hardware mute, wake-word performance and speech recognition **only after separate authorization**.

The first nine categories can be meaningfully tested offline with independent synthetic helpers. They cannot establish the tenth.

See [Architecture](architecture.md) for the parallel-processing trade-offs and [Build and integration](build-and-integration.md) for publication status and packaging gates.
