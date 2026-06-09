/* SPDX-License-Identifier: (GPL-2.0 OR BSD-3-Clause)
 *
 * Cross-arch register-file shim for the uprobe programs (U-021).
 *
 * bpf2go cross-compiles sslsniff for amd64 AND arm64 from one build host,
 * but vmlinux.h is dumped from that host's BTF: an x86_64-dumped vmlinux.h
 * defines x86's struct pt_regs and NOT arm64's struct user_pt_regs, the
 * type that bpf_tracing.h's PT_REGS_PARM1..N and PT_REGS_RC macros cast
 * the uprobe context to under __TARGET_ARCH_arm64. The definition below is
 * the arm64 UAPI register file (uapi/asm/ptrace.h) -- stable kernel ABI,
 * so defining it here is safe.
 *
 * Building the objects on an arm64 host (e.g. the arm64 kernel-matrix runner,
 * or an arm64 dev laptop)? Its vmlinux.h ALREADY defines struct user_pt_regs,
 * so this shim would redefine it and clang fails. The build must then pass
 * -DPROBECTL_VMLINUX_HAS_USER_PT_REGS to skip the shim. gen_bpf.sh does this
 * automatically: it greps the freshly-dumped vmlinux.h and sets the flag iff
 * the real struct is present -- so native-arm64, native-amd64, and the
 * amd64-host -> arm64 cross-build (release.yml) all compile from one rule. The
 * mirror case (an arm64 host cross-building an x86 object) would need an x86
 * struct pt_regs shim and is not done by any pipeline.
 */
#ifndef PROBECTL_BPF_ARCH_COMPAT_H
#define PROBECTL_BPF_ARCH_COMPAT_H

#if defined(__TARGET_ARCH_arm64) && !defined(PROBECTL_VMLINUX_HAS_USER_PT_REGS)
struct user_pt_regs {
	__u64 regs[31];
	__u64 sp;
	__u64 pc;
	__u64 pstate;
};
#endif

#endif /* PROBECTL_BPF_ARCH_COMPAT_H */
