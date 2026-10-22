# Owner-local Franken ASP tools

See [research documentation](../../docs/frankenasp/README.md).
Keep all extracted factory components in gitignored vendor-local/.
These tools do not enable the live native DSP by themselves.
Use prepare_runtime.sh only for owner-local staging; it does not install to an Echo.

For the current Athena branch, use `python3 install.py --backup-partitions DIR --device root@ECHO` for a read-only compatibility check; use `--install` **only** when deliberately staging your own ELF runtime and independently built worker/TECHO5 experimental binary. It does not activate the ASP. Full instructions: [installation.md](../../docs/frankenasp/installation.md).
