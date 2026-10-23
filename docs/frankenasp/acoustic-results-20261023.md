# Initial acoustic results — Echo Show 8 Gen 2 (Athena)

**23 October 2026 · Research preview · Wake-word capture, not loudspeaker audio quality**

The FrankenASP proof of concept is now running on real Athena hardware, using the device's original Amazon speech-input ASP through a private ARM32/Bionic worker. Startup, continuous PCM output, wake-word input, and guarded recovery have been demonstrated. Controlled, repeated acoustic testing shows a substantial initial wake-word detection advantage over the stock TECHO5 voice frontend. This is a practical engineering comparison, **not** a long-duration reliability qualification or an academic characterization of individual DSP algorithms.

## First-pass results

**Successful wake detections / 60 scheduled utterances** (near misses do *not* count as detections):

| Speech stimulus level | Stock TECHO5 mono | Stock TECHO5 stereo | FrankenASP mono | FrankenASP stereo |
| --- | ---: | ---: | ---: | ---: |
| 30 — quiet | Not in this matched series | Not in this matched series | **41/60 (68.3%)** | **45/60 (75.0%)** |
| 35 — normal | 45/60 (75.0%) | 22/60 (36.7%) | **59/60 (98.3%)** | **54/60 (90.0%)** |
| 40 — loud | 49/60 (81.7%) | 33/60 (55.0%) | **60/60 (100%)** | **59/60 (98.3%)** |

The **30/35/40 labels denote controlled stimulus playback levels**, used here as quiet/normal/loud test voice levels; they are not calibrated sound-pressure levels. An earlier stock TECHO5 level-30 session exists but is not merged into the later comparison matrix as an equivalent full control.

**Threshold caveat (important):** stock TECHO5 had MWW cutoff configured at 0.85, but the *observed effective cutoff during music playback* was **0.75**. FrankenASP's observed effective cutoff remained **0.85**. Thus the factory frontend's higher detection rate was observed despite its *stricter* wake-word threshold. This comparison describes end-to-end behavior under the measured settings; it is not a same-threshold DSP-only isolation experiment. False-positive trade-offs were not established by these trials.

## Test approach, in brief

- Same physical Athena Echo Show 8 Gen 2, wake-word model, room arrangement, and scripted speech material.
- Five predefined spatial scenarios per level, **12 utterances each** (60 per mode/level), including separately positioned voice and interfering TV audio.
- Fixed background music track, controlled restart/playback position, playback-state verification, and explicit mono/stereo selection.
- Wake events/near misses collected from device logs and matched to the stimulus schedule. Completed stock runs cover levels 35 and 40; FrankenASP was tested at 30, 35 and 40 with both speaker modes.
- The prototype uses four real microphone channels and the original factory speech-processing library. Amazon binaries, calibration data, private voice recordings and raw device logs are **not** distributed in this repository.

## What this establishes — and what it does not

Within these tested conditions, the factory ASP path consistently achieved **markedly better wake-word capture**, including adverse spatial/interference scenarios and the stereo-output condition. The results demonstrate the size of the current practical voice-input gap; they **do not identify AEC, beamforming, beam selection, denoising, or any particular component as its proven cause**.

The prototype does not reproduce the complete original Fire OS voice stack. Long-running stability, false-positive rate, cross-device and cross-firmware compatibility, and **post-wake command intelligibility/STT under spatial interference** remain for later research.

A more detailed account of the automated test rig, playback and microphone geometry, synchronized captures, controlled comparisons, spatial-tracking hypotheses and future design options will be published separately. These preliminary findings should not delay that work or be represented as a production-ready FrankenASP release.
