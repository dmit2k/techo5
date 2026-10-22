# Owner-local ARM32/Bionic runtime preparation

**Do not commit or redistribute a factory payload.** Obtain legally
appropriate FireOS `system` and `vendor` partition images from your
own Athena, and retain backup integrity evidence. The scripts only
read images and create files within the explicitly provided output
directory; they do not download binaries or touch the running Echo.

Requirements: Python 3, `debugfs`, `readelf` and (to build the worker)
the official Android NDK including the ARMv7 API 28 compiler.

~~~sh
cd tools/frankenasp
python3 extract_runtime.py   --system-image /path/to/your/p21-system.img   --vendor-image /path/to/your/p12-vendor.img   --output vendor-local/runtime
python3 extract_configs.py   --vendor-image /path/to/your/p12-vendor.img   --output vendor-local/runtime
export ANDROID_NDK_HOME=/path/to/android-ndk
./build_worker.sh
python3 -m unittest discover -s tests -v
~~~

`extract_runtime.py` walks a heuristic ELF dependency closure from
`/vendor/lib/libasp.so` and reports unresolved sonames. It is NOT a substitute
for verifying Android linker namespaces, symbol versions and ABI.
Vendor, VNDK and Bionic components must correspond to the factory build.
The output contains licensed proprietary material and is gitignored.

**Not yet a device installation procedure:** the old prototype used
an on-device `sh-native` ELF interpreter adaptation and relied on
FireOS vendor paths. Mounting/namespace layout, native linking, service
supervision and safe A/B fallback are yet to be made reproducible.

The opt-in marker records *intent*, not confirmed execution. The read-only
`status.sh` checks native ASP and worker shared objects in process maps,
but it does not yet certify real-time processing health.
