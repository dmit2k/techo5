# Controlled comparison test protocol (draft)

| Mode | Speaker path | Speech path |
|---|---|---|
| Official mono | mono | TECHO5 software AEC |
| Official stereo | stereo | TECHO5 software AEC |
| FrankenASP mono | mono | original native factory processor, physical AWB |
| FrankenASP stereo | stereo | original native factory processor, physical AWB |

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

The guarded Athena proof-of-concept passed short live startup,
worker-failure and stalled-IPC recovery tests. Controlled playback-based
**wake-word acoustic trials were completed on 23 October 2026**;
see the [initial concise results](acoustic-results-20261023.md).
These trials do not replace future synchronized four-microphone/AWB
recordings, post-wake STT comparisons, or long-duration AFE stability tests.

Follow the [Stage 3 preparation checklist](stage3-preparation-20261023.md)
before declaring any recorded clip suitable for AEC comparisons: in
particular, ALSA reference-frame delivery is not evidence of nonzero
playback-reference samples. The measured stock effective cutoff was 0.75
during music, while FrankenASP remained at 0.85. Record this explicitly
in every comparison; do not silently assume a common threshold.

