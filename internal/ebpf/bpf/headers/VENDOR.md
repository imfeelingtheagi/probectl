# Vendored libbpf BPF-program headers

These are the **BPF-side** headers from [libbpf](https://github.com/libbpf/libbpf),
vendored so the eBPF object compile (`bpf2go` → clang) does **not** depend on
whatever `libbpf-dev` version the build image or CI runner happens to ship.

Why: `sslsniff.bpf.c` uses `BPF_UPROBE`/`BPF_URETPROBE`, which libbpf only added
in **v1.2.0**. The shipped-agent build image (`golang:1.26-bookworm`) installs
libbpf **1.1.0**, which lacks those macros — so the compile failed. Vendoring a
pinned header set removes that whole class of version-skew breakage.

- Source: https://github.com/libbpf/libbpf
- Version: **v1.5.0**
- Commit: `09b9e83102eb8ab9e540d36b4559c55f3bcdb95d`
- Obtained via: `make install_headers` at that tag (then copied the BPF-program subset).
- License: LGPL-2.1 OR BSD-2-Clause (per each file's SPDX header).

Files (BPF-program headers only; the userspace loader headers are intentionally
NOT vendored — the Go side uses the `github.com/cilium/ebpf` library):

| file               | included by                          |
|--------------------|--------------------------------------|
| bpf_helpers.h      | l4flow.bpf.c, sslsniff.bpf.c         |
| bpf_helper_defs.h  | (via bpf_helpers.h)                  |
| bpf_tracing.h      | sslsniff.bpf.c (BPF_UPROBE)          |
| bpf_core_read.h    | available for CO-RE programs         |
| bpf_endian.h       | available for byte-order helpers     |

## Updating

Re-run `make install_headers` at the desired libbpf tag and copy the same five
files here; bump the Version/Commit above. Keep the floor at **>= v1.2.0**
(first release with BPF_UPROBE/BPF_URETPROBE).
