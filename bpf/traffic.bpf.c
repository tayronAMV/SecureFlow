#include "vmlinux.h"
#include "traffic.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

char LICENSE[] SEC("license") = "GPL";
    
struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 24); // 16MB buffer
} events SEC(".maps");

static __always_inline int parse_packet(struct __sk_buff *skb) {
    void *data = (void *)(long)skb->data;
    void *data_end = (void *)(long)skb->data_end;

    
    struct iphdr *ip = data;
    if ((void *)(ip + 1) > data_end)
        return 1;

    struct flow_event_t *event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
    if (!event)
        return 1;

    event->src_ip = ip->saddr;
    event->dst_ip = ip->daddr;
    event->protocol = ip->protocol;
    event->pkt_len = skb->len;

    if (ip->protocol == IPPROTO_TCP) {
        struct tcphdr *tcp = (void *)(ip + 1);
        if ((void *)(tcp + 1) <= data_end) {
            event->src_port = bpf_ntohs(tcp->source);
            event->dst_port = bpf_ntohs(tcp->dest);
        }
    } else if (ip->protocol == IPPROTO_UDP) {
        struct udphdr *udp = (void *)(ip + 1);
        if ((void *)(udp + 1) <= data_end) {
            event->src_port = bpf_ntohs(udp->source);
            event->dst_port = bpf_ntohs(udp->dest);
        }
    }

    bpf_ringbuf_submit(event, 0);
    return 1; // Accept packet
}

SEC("cgroup_skb/ingress")
int monitor_ingress(struct __sk_buff *skb) {
    return parse_packet(skb);
}

SEC("cgroup_skb/egress")
int monitor_egress(struct __sk_buff *skb) {
    return parse_packet(skb);
}
