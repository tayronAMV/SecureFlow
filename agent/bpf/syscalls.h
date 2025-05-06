#pragma once

#define EVENT_EXECVE    1
#define EVENT_EXECVEAT  2
#define EVENT_OPEN      3
#define EVENT_UNLINK    4
#define EVENT_CHMOD     5
#define EVENT_MOUNT     6
#define EVENT_SETUID    7
#define EVENT_SOCKET    8
#define EVENT_CONNECT   9

//set the types , in go code i dentify type of syscalls by the type 

struct syscall_event_t {
    u32 pid;           // PID of the process that made the syscall
    u32 type;          // Type of syscall (values 1–9 based on defined constants)
    char comm[TASK_COMM_LEN];     // Name of the process (from task_struct->comm)
    char filename[256]; // Target file or path used in the syscall (e.g., file opened, binary executed)    
};
