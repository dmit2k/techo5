# Controlled comparison test protocol (draft)

| Mode | Speaker path | Speech path |
|---|---|---|
| Official mono | mono | TECHO5 software AEC |
| Official stereo | stereo | TECHO5 software AEC |
| Franken ASP | stereo | original native processor, physical AWB |

Hold constant wake model, threshold, sound level, room positions, playback
content, background and STT pipeline. Report true positives, false
positives, false negatives, near misses, latencies and clipped word onsets.
Transcribe short commands for WER/CER, and vary speaking distance, voice
level, phrasing, TV/music interference and volume in documented steps.

Each session records Git commit, actual runtime binary, DSP backend
**confirmed active**, MWW model and threshold, AEC reference origin,
channel mapping, whether IDME calibration applied (NOT coefficient
values), timestamps and clock basis. Include CPU, memory, PCM XRUN,
lost frames, watchdog events and long-duration playback failures.

When physical trials cannot be simultaneous, describe confounders;
do not claim that different loudness or RMS alone establishes superior
AEC. Publish aggregates and only synthetic or explicitly consented
recordings. Never share unredacted device logs, private audio, IDME,
factory coefficients or original software images.

The guarded Athena proof-of-concept has passed short live startup,
worker-failure and stalled-IPC recovery tests. No further live audio
experiments or long-duration AFE stability test are planned in this phase.
The reproducible offline A/B comparison is **deferred until after the
scheduled acoustic capture session**, which will provide more labeled
four-microphone and hardware-reference recordings.

Follow the [Stage 3 preparation checklist](stage3-preparation-20261023.md)
before declaring any recorded clip suitable for AEC comparisons: in
particular, ALSA reference-frame delivery is not evidence of nonzero
playback-reference samples. Keep the stock cutoff 0.75 and fixed 0.85
as **separate documented conditions**, never an unlogged setting change.

