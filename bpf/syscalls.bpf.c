// bpf/syscalls.bpf.c
// eBPF LSM and tracepoint hooks for nine critical syscalls

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_core_read.h>
#include "syscalls.h"

// Ring buffer map for syscall events
struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 24);
} syscall_events SEC(".maps");

// -----------------------------
// EXECUTION: execve via LSM
// -----------------------------
SEC("lsm/bprm_check")
int log_execve(struct linux_binprm *bprm)
{
    struct syscall_event_t event = {};
    u64 id = bpf_get_current_pid_tgid();
    event.pid = id >> 32;
    event.type = EVENT_EXECVE;
    bpf_get_current_comm(&event.comm, sizeof(event.comm));
    bpf_probe_read_user_str(event.filename, sizeof(event.filename), bprm->filename);
    bpf_ringbuf_output(&syscall_events, &event, sizeof(event), 0);
    return 0;
}

// -----------------------------
// EXECUTION: execveat via tracepoint
// -----------------------------
SEC("tracepoint/syscalls/sys_enter_execveat")
int log_execveat(struct trace_event_raw_sys_enter *ctx)
{
    struct syscall_event_t event = {};
    u64 id = bpf_get_current_pid_tgid();
    event.pid = id >> 32;
    event.type = EVENT_EXECVEAT;
    bpf_get_current_comm(&event.comm, sizeof(event.comm));
    const char *pathname = (const char *)ctx->args[1];
    bpf_probe_read_user_str(event.filename, sizeof(event.filename), pathname);
    bpf_ringbuf_output(&syscall_events, &event, sizeof(event), 0);
    return 0;
}

// -----------------------------
// FILE OPERATIONS: open via LSM
// -----------------------------
// LSM hook file_open has two params: struct file *file, const struct cred *ctx
SEC("lsm/file_open")
int log_open(struct file *file, const struct cred *ctx)
{
    struct syscall_event_t event = {};
    u64 id = bpf_get_current_pid_tgid();
    event.pid = id >> 32;
    event.type = EVENT_OPEN;
    bpf_get_current_comm(&event.comm, sizeof(event.comm));
    // file->f_path.dentry->d_name.name is a char *
    bpf_probe_read_kernel_str(event.filename, sizeof(event.filename), 
        file->f_path.dentry->d_name.name);
    bpf_ringbuf_output(&syscall_events, &event, sizeof(event), 0);
    return 0;
}

// -----------------------------
// FILE OPERATIONS: unlink via LSM
// -----------------------------
SEC("lsm/inode_unlink")
int log_unlink(struct inode *dir, struct dentry *dentry)
{
    struct syscall_event_t event = {};
    u64 id = bpf_get_current_pid_tgid();
    event.pid = id >> 32;
    event.type = EVENT_UNLINK;
    bpf_get_current_comm(&event.comm, sizeof(event.comm));
    bpf_probe_read_kernel_str(event.filename, sizeof(event.filename), 
        dentry->d_name.name);
    bpf_ringbuf_output(&syscall_events, &event, sizeof(event), 0);
    return 0;
}

// -----------------------------
// FILE PERMISSIONS: chmod via tracepoint
// -----------------------------
SEC("tracepoint/syscalls/sys_enter_chmod")
int log_chmod(struct trace_event_raw_sys_enter *ctx)
{
    struct syscall_event_t event = {};
    u64 id = bpf_get_current_pid_tgid();
    event.pid = id >> 32;
    event.type = EVENT_CHMOD;
    bpf_get_current_comm(&event.comm, sizeof(event.comm));
    const char *path = (const char *)ctx->args[0];
    bpf_probe_read_user_str(event.filename, sizeof(event.filename), path);
    bpf_ringbuf_output(&syscall_events, &event, sizeof(event), 0);
    return 0;
}

// -----------------------------
// MOUNT: mount via tracepoint
// -----------------------------
SEC("tracepoint/syscalls/sys_enter_mount")
int log_mount(struct trace_event_raw_sys_enter *ctx)
{
    struct syscall_event_t event = {};
    u64 id = bpf_get_current_pid_tgid();
    event.pid = id >> 32;
    event.type = EVENT_MOUNT;
    bpf_get_current_comm(&event.comm, sizeof(event.comm));
    const char *target = (const char *)ctx->args[1];
    bpf_probe_read_user_str(event.filename, sizeof(event.filename), target);
    bpf_ringbuf_output(&syscall_events, &event, sizeof(event), 0);
    return 0;
}

// -----------------------------
// PRIVILEGE ESCALATION: setuid via tracepoint
// -----------------------------
SEC("tracepoint/syscalls/sys_enter_setuid")
int log_setuid(struct trace_event_raw_sys_enter *ctx)
{
    struct syscall_event_t event = {};
    u64 id = bpf_get_current_pid_tgid();
    event.pid = id >> 32;
    event.type = EVENT_SETUID;
    bpf_get_current_comm(&event.comm, sizeof(event.comm));
    bpf_ringbuf_output(&syscall_events, &event, sizeof(event), 0);
    return 0;
}

// -----------------------------
// NETWORK: socket via tracepoint
// -----------------------------
SEC("tracepoint/syscalls/sys_enter_socket")
int log_socket(struct trace_event_raw_sys_enter *ctx)
{
    struct syscall_event_t event = {};
    u64 id = bpf_get_current_pid_tgid();
    event.pid = id >> 32;
    event.type = EVENT_SOCKET;
    bpf_get_current_comm(&event.comm, sizeof(event.comm));
    bpf_ringbuf_output(&syscall_events, &event, sizeof(event), 0);
    return 0;
}

// -----------------------------
// NETWORK: connect via tracepoint
// -----------------------------
SEC("tracepoint/syscalls/sys_enter_connect")
int log_connect(struct trace_event_raw_sys_enter *ctx)
{
    struct syscall_event_t event = {};
    u64 id = bpf_get_current_pid_tgid();
    event.pid = id >> 32;
    event.type = EVENT_CONNECT;
    bpf_get_current_comm(&event.comm, sizeof(event.comm));
    bpf_ringbuf_output(&syscall_events, &event, sizeof(event), 0);
    return 0;
}

char LICENSE[] SEC("license") = "GPL";
