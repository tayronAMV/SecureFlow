#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

SEC("tc_ingress")
int monitor_ingress(struct __sk_buff *ctx) {
    char msg[] = "🔥 ingress hit\\n";
    bpf_trace_printk(msg, sizeof(msg));
    return 0;
}

char LICENSE[] SEC("license") = "GPL";