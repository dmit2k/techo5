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

No on-device trials until a tested official-binary recovery path exists.
