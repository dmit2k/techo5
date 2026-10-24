# Optional 15-second wake/STT PCM clips — instrumented stock TECHO5 v1.6.0

This source branch starts from the **unaltered official upstream v1.6.0 commit `6c2828192ec58fd930977f86ef7f9a322fe7c3c4`** and adds only research clip capture plus the constant processing-mode label `techo5`. It does not contain the FrankenASP glue. It is **not an official TECHO5 binary or release**.

## Behaviour

- Ordinary uninstrumented behavior: `Keep wake clips` stores the four seconds **before** a wake/near miss.
- Opt-in behavior: a marker file present before `echod` starts extends event clips to **four seconds before + eleven seconds after**. The recorder takes the same **16-kHz mono PCM** that this stock speech-processing path hands to wake-word and STT.
- The regular `wake_clips` / `wake_clip_audio` actions continue working. The WAV/JSON pairs appear under the standard `wake-clips` directory and are limited to the existing 100 files per kind.
- JSON `processing_mode` is `techo5` and research clips include `duration_seconds`.

## Activation

With owner authorization, use the established backup/rollback and application installation procedure for a build produced from this branch. **No original Amazon binaries are needed.** The feature is disabled unless the following marker exists before startup:

```sh
touch /data/misc/techo5/wake-clip-15s.enabled
```

The ordinary Home Assistant `Keep wake clips` setting must also be ON. A daemon restart through the supported mechanism is required to read the marker. To return to 4-second clips, delete the marker and restart the instrumented daemon:

```sh
rm -- /data/misc/techo5/wake-clip-15s.enabled
```

**Disabling `Keep wake clips` deletes retained wake clips.** Keep private files in an authorized owner-local location, not GitHub. Fully populated 15-second WAV retention at 100 clips × 4 categories uses about 192 MB; check free space.

## Comparing with FrankenASP

Use identical wake-model threshold, TV soundtrack/volume, listener orientation/distance, source speech and other settings. Record several trials of wake + command, wake + silence, and TV alone, without publishing voice recordings. The other experiment branch, `experiment/athena-15s-voice-clips`, records `processing_mode: factory_asp` when the factory pipeline is *actually active*, and `techo5` otherwise.

Do **not** assume that absence of transcription alone proves a DSP beam direction or speaker identity. Listening to both actual processed mono WAV streams is necessary. This branch differs from the FrankenASP fork in more than the DSP selection because the latter started at a newer v1.6.0-era upstream commit; comparisons using the exact same FA-capable binary switched to stock processing are a useful additional control.

