# External ASP V3 — clean restart documents

**Status:** CORRECTIVE DRAFT — owner's acceptance and Git cleanup authorisation pending.
**Frozen source bases:**
- TECHO5 main: 62603a7854283a3a07473a17c831c6ef0fd6f8d4
- FrankenASP V3 main (must be restored before Stage01): e986ae18643e08e95f29a38aae1dd74762f6362e

## Authority and reading order

1. REGISTER_58_DECISIONS.md — existing numbered decisions, corrected to raw byte contract.
2. FINAL_FIVE_CLARIFICATIONS.md — boundary ownership clarified.
3. SPEC_V2_APPROVED.md — existing normative filename, but **V3 restart corrected draft not yet re-approved**; original V2 audit does not approve new changes.
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

This folder is a **prepared non-repository copy** of the 24 corrected canonical documents for a future single clean-base documentation commit. No prior stage handoffs or failed implementation code are carried into the new docs history.

DO NOT execute Stage01, remove any Git ref/worktree, reset, force-push, or deploy to Echo until Git cleanup plan and corrected documents are separately accepted by the owner. Source old stage code is reference only.

After owner authorises cleanup: recreate docs branch from clean TECHO5 main and commit these corrected documents once. Then record real docs commit SHA and updated SPEC blob in execution handoffs.
