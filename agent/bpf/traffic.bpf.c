// traffic.bpf.c – verifier‑friendly version matching the slim flow_event_t
// Only HTTP, DNS and ICMP DPI are supported.  No helper calls that are
// forbidden in cgroup_skb programs (bpf_probe_read_kernel removed).
// Loops are fully unrolled with constant bounds to satisfy the verifier.

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <traffic.h>
#include <bpf/bpf_endian.h> 


#define TCP 6
#define UDP 17
#define ICMP 1
#define ETH_P_IP 0x0800

struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 24);
} events SEC(".maps");

static __always_inline int check_bounds(void *ptr, void *data_end, __u64 size)
{
    return ptr + size > data_end;
}

/* copy at most max_len bytes from payload into buf using bpf_skb_load_bytes */
static __always_inline int load_payload(struct __sk_buff *ctx, void *data, void *payload,
                                        void *data_end, char *buf, int max_len)
{
    int avail = (int)((long)data_end - (long)payload);
    int to_copy = avail < max_len ? avail : max_len;
    if (to_copy <= 0)
        return 0;
    if (bpf_skb_load_bytes(ctx, (long)payload - (long)data, buf, to_copy) < 0)
        return 0;
    return to_copy;
}

static __always_inline void parse_http(struct __sk_buff *ctx, void *data, void *payload,
                                       void *data_end, struct flow_event_t *evt)
{
    char buf[256] = {};
    int len = load_payload(ctx, data, payload, data_end, buf, sizeof(buf));

    /* Parse HTTP method (max 8 bytes) */
#pragma unroll
    for (int i = 0; i < 8; i++) {
        if (i >= len || buf[i] == ' ') break;
        evt->http.method[i] = buf[i];
    }

    /* Parse path (max 64 bytes) */
    int p_off = 0;
#pragma unroll
    for(int i = 0; i < 8; i++) {
        if (i >= len || buf[i] == ' ') { p_off = i + 1; break; }
    }
#pragma unroll
    for (int i = 0; i < 64; i++) {
        if (p_off + i >= len || buf[p_off + i] == ' ') break;
        evt->http.path[i] = buf[p_off + i];
    }

    evt->dpi_protocol = 1; /* HTTP */
}

static __always_inline void parse_dns(struct __sk_buff *ctx, void *data, void *payload,
                                      void *data_end, struct flow_event_t *evt)
{
    /* DNS header is 12 bytes, question section follows */
    void *q = payload + 12;
    if (q >= data_end) return;

    char name[64] = {};
    int len = load_payload(ctx, data, q, data_end, name, sizeof(name));

#pragma unroll
    for (int i = 0; i < 64; i++) {
        if (i >= len) break;
        char c = name[i];
        evt->dns.query_name[i] = c;
        if (c == 0) { len = i + 1; break; }
    }

    /* query_type follows the name + null */
    if ((void *)q + len + 2 <= data_end) {
        __u16 qtype = 0;
        bpf_skb_load_bytes(ctx, (long)((void *)q - data) + len, &qtype, sizeof(qtype));
        evt->dns.query_type = bpf_ntohs(qtype);
    }

    evt->dpi_protocol = 2; /* DNS */
}

static __always_inline void parse_icmp(struct icmphdr *icmp, struct flow_event_t *evt)
{
    evt->icmp.icmp_type = icmp->type;
    evt->dpi_protocol  = 3; /* ICMP */
}

static __always_inline int emit_and_return(struct flow_event_t *evt, int verdict)
{
    bpf_ringbuf_submit(evt, 0);
    return verdict; /* 0 = accept */
}

static __always_inline int drop_and_return(struct flow_event_t *evt, int verdict)
{
    bpf_ringbuf_discard(evt, 0);
    return verdict;
}

static __always_inline int parse_packet(struct __sk_buff *ctx, __u8 direction)
{
    void *data     = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;

    struct flow_event_t *evt = bpf_ringbuf_reserve(&events, sizeof(*evt), 0);
    if (!evt) return 0;

    __builtin_memset(evt, 0, sizeof(*evt));

    evt->timestamp   = bpf_ktime_get_ns();
    evt->payload_len = ctx->len;
    evt->direction   = direction;
    evt->Pid = bpf_get_current_pid_tgid() >> 32;

    struct ethhdr *eth = data;
    if (check_bounds(eth, data_end, sizeof(*eth)))
        return drop_and_return(evt, 0);

    if (bpf_ntohs(eth->h_proto) != ETH_P_IP)
        return drop_and_return(evt, 0);

    struct iphdr *ip = (void *)(eth + 1);
    if (check_bounds(ip, data_end, sizeof(*ip)))
        return drop_and_return(evt, 0);

    evt->src_ip  = ip->saddr;
    evt->dst_ip  = ip->daddr;
    evt->protocol = ip->protocol;

    void *l4 = (void *)ip + (ip->ihl * 4);
    if (l4 > data_end) return drop_and_return(evt, 0);

    if (ip->protocol == TCP) {
        struct tcphdr *tcp = l4;
        if (check_bounds(tcp, data_end, sizeof(*tcp)))
            return drop_and_return(evt, 0);

        evt->src_port = bpf_ntohs(tcp->source);
        evt->dst_port = bpf_ntohs(tcp->dest);

        void *payload = (void *)tcp + (tcp->doff * 4);
        if (payload < data_end && evt->dst_port == 80)
            parse_http(ctx, data, payload, data_end, evt);

        return emit_and_return(evt, 0);

    } else if (ip->protocol == UDP) {
        struct udphdr *udp = l4;
        if (check_bounds(udp, data_end, sizeof(*udp)))
            return drop_and_return(evt, 0);

        evt->src_port = bpf_ntohs(udp->source);
        evt->dst_port = bpf_ntohs(udp->dest);

        void *payload = (void *)udp + sizeof(*udp);
        if (payload < data_end && (evt->dst_port == 53 || evt->src_port == 53))
            parse_dns(ctx, data, payload, data_end, evt);

        return emit_and_return(evt, 0);

    } else if (ip->protocol == ICMP) {
        struct icmphdr *icmp = l4;
        if (check_bounds(icmp, data_end, sizeof(*icmp)))
            return drop_and_return(evt, 0);

        parse_icmp(icmp, evt);
        return emit_and_return(evt, 0);
    }

    /* Other protocols – discard event */
    return drop_and_return(evt, 0);
}

SEC("cgroup_skb/ingress")
int monitor_ingress(struct __sk_buff *ctx)
{
    return parse_packet(ctx, 1);
}

SEC("cgroup_skb/egress")
int monitor_egress(struct __sk_buff *ctx)
{
    return parse_packet(ctx, 0);
}

char LICENSE[] SEC("license") = "GPL";
