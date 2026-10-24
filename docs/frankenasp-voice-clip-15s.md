# Research: 15-second voice-path comparison clips (Athena)

This is **optional instrumentation**, not a change to the archived FrankenASP/TECHO5 v1.6.0-era baseline, the stock Amazon DSP libraries, or the microphone array. No device is modified by compiling this branch.

## What is recorded

With the opt-in marker, `Keep wake clips` saves a maximum **15-second, 16 kHz / PCM16 / mono WAV** for each qualifying wake or near miss: **up to 4 seconds before** the scoring moment and **11 seconds after**. This is the processed speech PCM heard by the wake engine and sent to the STT path; **not** six channels of raw audio. In normal mode the original four-second pre-event clip behavior is retained.

The 15-second tail is collected even after the wake was already classified. An initial period shorter than four seconds is possible immediately after starting or enabling recording. If the mic is muted, a phone/intercom/talkback claims audio, recording is switched off, or the stop word fires, an unfinished long clip is discarded rather than joining audio across that transition.

Each WAV has a JSON sidecar with `processing_mode` (`factory_asp` or `techo5`, observed from the actual mono PCM path), and `duration_seconds` for study clips, in addition to the original timestamp, model, kind, peak and cutoff. `Keep wake clips` must be enabled in the device settings. Turn it off to delete all kept wake clips, as in normal TECHO5.

## Enable on an authorized test unit

The experimental 15-second recorder activates **only at daemon startup** when this marker exists:

```sh
touch /data/techo5-linux/wake-clip-15s.enabled
```

Deploy the **instrumented** `echod` from this experiment branch using your existing safe staging/rollback procedure, then restart the daemon through the supported device mechanism. This marker does nothing in the unmodified official binary. Do not overwrite a protected or baseline binary without a rollback image. To disable extended recording, remove the marker and restart the instrumented daemon; ordinary four-second recording remains available.

```sh
rm -- /data/techo5-linux/wake-clip-15s.enabled
```

The resulting WAVs and JSON sidecars are under `/data/techo5-linux/wake-clips/{wake,nearmiss,mistake,noise}/`. The usual Home Assistant actions `wake_clips` and `wake_clip_audio` also expose them. Audio files are private; **do not commit, publish, or upload real speech recordings**.

## A/B procedure

Use the **same instrumented `echod` binary and settings**, and alternate its source of speech PCM:

- **FA run**: explicitly arm the factory ASP using the existing guarded FA mechanism. Confirm the new clip JSON says `processing_mode: factory_asp`.
- **TECHO5 run**: disable FA via the existing guarded mechanism and confirm the returned-to-stock processing path. Confirm JSON says `processing_mode: techo5`. This is a TECHO5 processing baseline *inside the instrumented fork*, not a claim that its binary is byte-for-byte identical to an official release.

Keep mic model/threshold, TV program and approximate level, distance, orientation, output volume, and voice phrases comparable. Make several wake + command trials and wake + silence trials with the TV on; include TV-only trials. Preserve WAV and JSON pairs, but keep them owner-local. A 2-second follow-up timer alone does not establish 15 seconds of STT transcription; the clip captures processed mono regardless of that timeout.

**Limits:** neither a quiet WAV nor the absence of an STT transcript proves the vendor DSP locks onto a voice identity. Beamforming, VAD, noise suppression and downstream STT all remain possible contributors. These captures let you evaluate the processed PCM directly. Mode transitions must use the existing safe FA controls and **must not be inferred from the marker alone**.

## Storage budget

A full 15-second WAV is about **480 KB** (16,000 samples/s × 2 bytes × 15 s), vs 128 KB at four seconds. The existing `clipsKept = 100` per each of four categories can therefore retain up to **approximately 192 MB** in WAV data; plan adequate `/data` free space. Concurrent unfinished captures are capped at eight. This is experimental and is intentionally opt-in.

