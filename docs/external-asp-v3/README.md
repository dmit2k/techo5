# External ASP V3 — clean restart documents

**Status:** OWNER ACCEPTED 2026-10-31; Stage01 execution PASS / acceptance ACCEPTED / integration COMPLETE. Stage02 requires its own execution assignment.
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
- Factory IDME/hardware metadata go to helper as original bytes. No Go layout.Idme()/parseIdme() derived raw values, Q30, gain masks, availability/confidence, captured MIC slot association or DSP-specific qualification.
- FrankenASP alone interprets raw values, maps channels, performs DSP-specific conversions, and may refuse unsupported inputs before READY.
- New Go and C integration lines are respectively experimental/external-asp-v3-dev-main and experimental/v3-dev-main. **Neither main moves during Stage01–07.**
- Stage02 executor chooses minimal Go/C raw metadata encoding and startup handshake details; no additional framework or architecture is mandated.

## Git and execution status

Cleanup and frozen-base documentation commit `06bb7e8a56ffd0406626d4f218f7dc796d799d03` completed. Stage01 evidence is in `stages/01_HANDOFF.md`, with owner acceptance recorded by commit `79f0803b38bf39de3362075f153600c26dff1b65`. Both main and experimental dev-main remain frozen at the source bases above. The corrected SPEC blob is `c754947d81c77860f4221135b84956249f956935`.

Stage02–07 require separately authorized execution and acceptance/integration. No new FF, reset or force-push; no Echo deployment or hardware tests. Historical rejected code is reference-only.
