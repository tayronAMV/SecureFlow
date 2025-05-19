#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_endian.h>
#include "traffic.h"

#define TCP 6
#define UDP 17
#define ICMP 1
#define ETH_P_IP 0x0800

struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 24);
} events SEC(".maps");

static __always_inline int check_bounds(void *ptr, void *data_end, __u64 size) {
    return ((char *)ptr + size) > (char *)data_end;
}

static __always_inline int load_payload(struct __sk_buff *ctx, void *data, void *payload,
                                        void *data_end, char *buf, int max_len) {
    int avail = (int)((long)data_end - (long)payload);
    int to_copy = avail < max_len ? avail : max_len;
    if (to_copy <= 0)
        return 0;
    if (bpf_skb_load_bytes(ctx, (long)payload - (long)data, buf, to_copy) < 0)
        return 0;
    return to_copy;
}

static __always_inline void parse_http(struct __sk_buff *ctx, void *data, void *payload,
                                       void *data_end, struct flow_event_t *evt) {
    char buf[256] = {};
    int len = load_payload(ctx, data, payload, data_end, buf, sizeof(buf));

#pragma unroll
    for (int i = 0; i < 8; i++) {
        if (i >= len) break;
        if (i >= len || buf[i] == ' ') break;
        evt->method[i] = buf[i];
    }

    int p_off = 0;
#pragma unroll
    for (int i = 0; i < 8; i++) {
        if (i >= len) break;
        if (i >= len || buf[i] == ' ') { p_off = i + 1; break; }
    }

#pragma unroll
    for (int i = 0; i < 64; i++) {
        if (i >= len) break;
        if (p_off + i >= len || buf[p_off + i] == ' ') break;
        evt->path[i] = buf[p_off + i];
    }

    evt->dpi_protocol = 1;
}

static __always_inline void parse_dns(struct __sk_buff *ctx, void *data, void *payload,
                                      void *data_end, struct flow_event_t *evt) {
    void *q = payload + 12;
    if (q >= data_end) return;

    char name[64] = {};
    int len = load_payload(ctx, data, q, data_end, name, sizeof(name));

#pragma unroll
    for (int i = 0; i < 64; i++) {
        if (i >= len) break;
        char c = name[i];
        evt->query_name[i] = c;
        if (c == 0) { len = i + 1; break; }
    }

    if ((void *)q + len + 2 <= data_end) {
        __u16 qtype = 0;
        bpf_skb_load_bytes(ctx, (long)((void *)q - data) + len, &qtype, sizeof(qtype));
        evt->query_type = bpf_ntohs(qtype);
    }

    evt->dpi_protocol = 2;
}

static __always_inline void parse_icmp(struct icmphdr *icmp, struct flow_event_t *evt) {
    evt->icmp_type = icmp->type;
    evt->dpi_protocol = 3;
}

static __always_inline int emit_and_return(struct flow_event_t *evt)
{
    bpf_printk("Submitting event: proto \n ");
    bpf_ringbuf_submit(evt, 0);
    return 1;                          // allow packet (BPF_SKB_PASS)
}

static __always_inline int discard_and_return(struct flow_event_t *evt)
{
        
    bpf_printk("Submitting event: proto \n ");
    bpf_ringbuf_submit(evt, 0);
    return 1;                          // still allow packet
}

static __always_inline int parse_packet(struct __sk_buff *ctx, __u8 direction) {
    void *data     = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;

    struct ethhdr *eth = data; // layer 2 parsing 
    if (check_bounds(eth, data_end, sizeof(*eth)))
        return 1;


    struct iphdr *ip = (void *)eth + sizeof(*eth); // layer 3 parsing 
    if (check_bounds(ip, data_end, sizeof(*ip)))
        return 1;


    // if (ip->ihl < 5 || (void *)ip + ip->ihl * 4 > data_end)
    //         return 1;

    /* only TCP / UDP / ICMP – otherwise just pass */
    if (!(ip->protocol == TCP || ip->protocol == UDP || ip->protocol == ICMP))
         return 1;

    /* allocate event */
    struct flow_event_t *evt = bpf_ringbuf_reserve(&events, sizeof(*evt), 0);
    if (!evt)
        return 0;


    void *l4 = (void *)ip + ip->ihl * 4; // layer 4 parsing ! 
        
    __builtin_memset(evt, 0, sizeof(*evt));
    evt->timestamp = bpf_ktime_get_ns();
    evt->payload_len = ctx->len;
    evt->direction = direction;
    evt->Pid = bpf_get_current_pid_tgid() >> 32;
    evt->src_ip = ip->saddr;
    evt->dst_ip = ip->daddr;
    evt->protocol = ip->protocol;

    if (ip->protocol == TCP) {
        struct tcphdr *tcp = l4;
       if (check_bounds(tcp, data_end, sizeof(*tcp)))
         return discard_and_return(evt);

        if (tcp->doff < 5 || (void *)tcp + tcp->doff * 4 > data_end)
            return discard_and_return(evt);

        evt->src_port = bpf_ntohs(tcp->source);
        evt->dst_port = bpf_ntohs(tcp->dest);

        void *payload = (void *)tcp + tcp->doff * 4;
        parse_http(ctx, data, payload, data_end, evt);
        return emit_and_return(evt);

    } else if (ip->protocol == UDP) {
        struct udphdr *udp = l4;
        if (check_bounds(udp, data_end, sizeof(*udp)))
            return discard_and_return(evt);

        evt->src_port = bpf_ntohs(udp->source);
        evt->dst_port = bpf_ntohs(udp->dest);
        void *payload = (void *)udp + sizeof(*udp);
        parse_dns(ctx, data, payload, data_end, evt);  // parse_dns will decide if it's really DNS

        return emit_and_return(evt);

    } else if (ip->protocol == ICMP) {
        struct icmphdr *icmp = l4;
        if (check_bounds(icmp, data_end, sizeof(*icmp)))
            return discard_and_return(evt);

        parse_icmp(icmp, evt);
        return emit_and_return(evt);
    }

    return discard_and_return(evt);
}

SEC("cgroup_skb/ingress")
int monitor_ingress(struct __sk_buff *ctx) {
    return parse_packet(ctx, 1);
}

SEC("cgroup_skb/egress")
int monitor_egress(struct __sk_buff *ctx) {
    return parse_packet(ctx, 0);
}

char LICENSE[] SEC("license") = "GPL";
