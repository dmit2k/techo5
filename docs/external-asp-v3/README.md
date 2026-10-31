# External ASP V3 — clean restart documents

**Status:** PCM-ONLY ARCHITECTURE CORRECTION, OWNER DIRECTED 2026-10-31. Historical Stage02/03 accepted pair T03/C02 uses obsolete metadata wire; correction Stage02-R→Stage03-R required before Stage04. See `PCM_ONLY_ARCHITECTURE_DECISION.md`.
**Frozen source bases:**
- TECHO5 main: 62603a7854283a3a07473a17c831c6ef0fd6f8d4
- FrankenASP V3 main (restored and frozen): e986ae18643e08e95f29a38aae1dd74762f6362e

## Authority and reading order

1. REGISTER_58_DECISIONS.md — existing numbered decisions, corrected to raw byte contract.
2. FINAL_FIVE_CLARIFICATIONS.md — boundary ownership clarified.
3. SPEC_V2_APPROVED.md — existing normative filename, but **V3 restart corrected contract accepted by the owner on 2026-10-31**; original V2 audit does not approve new changes.
4. MASTER_CONTRACT.md — shared minimal rules and two protected main branches.
5. EXECUTION_BRANCH_STRATEGY.md — two independent experimental integration branches.
6. DECOMPOSITION_PLAN.md, REQUIREMENTS_TRACEABILITY.md, HANDOFF_TEMPLATE.md — seven-stage dependencies and verification.
7. stages/01_BASELINE_BOOTSTRAP_TASK.md through stages/07_SOFTWARE_ACCEPTANCE_TASK.md, with one matching *_PROMPT.md per task.

AUDIT_V2_FINAL.md is preserved as an **historical audit of the pre-restart V2**. It is not evidence that the new raw IDME/REF and two-dev-main corrections have passed audit.

## Hard boundaries

- Host TECHO5 sends native MIC and REF unmodified, including **both** hardware stereo REF L/R samples, order, rate, format and interleaving; **only one SoftwareCut** may alter External-input PCM.
- External IPC is **PCM-only**: actual native MIC/REF bytes plus exact physical geometry, with no factory IDME, metadata or platform identity sections. Existing stock-only IDME reads stay stock. FrankenASP launcher separately prepares vendor resource views/IDME aliases before C-child; vendor Calibrator owns calibration.
- FrankenASP alone interprets raw values, maps channels, performs DSP-specific conversions, and may refuse unsupported inputs before READY.
- New Go and C integration lines are respectively experimental/external-asp-v3-dev-main and experimental/v3-dev-main. **Neither main moves during Stage01–07.**
- Correction Stage02 replaces the old metadata-bearing 24-byte header with the fixed 16-byte PCM-only header; agreed READY byte remains 0xA5. No dual-format compatibility or added framework.

## Git and execution status

Cleanup and frozen-base documentation commit `06bb7e8a56ffd0406626d4f218f7dc796d799d03` completed. Stage01 evidence is in `stages/01_HANDOFF.md`, with owner acceptance recorded by commit `79f0803b38bf39de3362075f153600c26dff1b65`. Both protected main refs remain at the frozen clean bases; dev-main refs advanced through accepted Stage02/03 to T03/C02 and are **not** yet PCM-only. The new owner-directed PCM-only correction requires Stage02-R/Stage03-R gates; current SPEC blob is pinned in MASTER_CONTRACT.

Stage02–07 require separately authorized execution and acceptance/integration. No new FF, reset or force-push; no Echo deployment or hardware tests. Historical rejected code is reference-only.
