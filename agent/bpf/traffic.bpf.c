#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>
#include "traffic.h"

#define TCP 6
#define UDP 17
#define ICMP 1
#define ETH_P_IP 0x0800

// TC return codes
#define TC_ACT_OK 0
#define TC_ACT_SHOT 2


struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 24);
} events SEC(".maps");


struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, 128);
    __type(key, __u32);
    __type(value, struct flow_rule_t);
} flow_rules SEC(".maps");




static __always_inline int match_rule(struct flow_event_t *event) {
    struct flow_rule_t *rule;
    for (int i = 0; i < 10; i++) {
        int count = 0;
        rule = bpf_map_lookup_elem(&flow_rules, &(int){i});
        if (!rule)
            continue;

        if (rule->action == 0)
            continue;

        if (rule->src_ip == event->src_ip) {
            bpf_printk("src_ip match: %x", event->src_ip);
            count++;
        }

        if (rule->dst_ip == event->dst_ip) {
            bpf_printk("dst_ip match: %x", event->dst_ip);
            count++;
        }

        if (rule->protocol == event->protocol) {
            bpf_printk("protocol match: %d", event->protocol);
            count++;
        }

        if (rule->src_port && rule->src_port == event->src_port) {
            bpf_printk("src_port match: %d", event->src_port);
            count++;
        }

        if (rule->dst_port && rule->dst_port == event->dst_port) {
            bpf_printk("dst_port match: %d", event->dst_port);
            count++;
        }

        
        if (rule->direction == event->direction) {
            bpf_printk("direction match: %d , %d ", event->direction ,rule->direction );
            count++;
        }

        if (rule->dpi_protocol && rule->dpi_protocol == event->dpi_protocol) {
            bpf_printk("dpi_protocol match: %d", event->dpi_protocol);
            count++;
        }

        if (rule->icmp_type == event->icmp_type) {
            bpf_printk("icmp_type match: %d", event->icmp_type);
            count++;
        }

        if (rule->query_type == event->query_type) {
            bpf_printk("query_type match: %d", event->query_type);
            count++;
        }

        if (rule->method[0] != 0 &&
            __builtin_memcmp(rule->method, event->method, 8) == 0) {
            bpf_printk("method match: %x", event->method[0]); // print first byte
            count++;
        }

        if (rule->path[0] != 0 &&
            __builtin_memcmp(rule->path, event->path, 64) == 0) {
            bpf_printk("path match: %x", event->path[0]); // print first byte
            count++;
        }

        if (rule->query_name[0] != 0 &&
            __builtin_memcmp(rule->query_name, event->query_name, 64) == 0) {
            bpf_printk("query_name match: %x", event->query_name[0]); // print first byte
            count++;
        }

        bpf_printk("total matches: %d of required %d", count, rule->action);

        if (count == rule->action) {
            bpf_printk("DROP");
            return 1;
        }
    }

    return 0; // Allow
}


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

    // Extract HTTP method
    for (int i = 0; i < 8 && i < len; i++) {
        if (buf[i] == ' ') break;
        evt->method[i] = buf[i];
    }
    
    // Find path start (after first space)
    int p_off = 0;
    for (int i = 0; i < 256 && i < len; i++) {
        if (buf[i] == ' ') { p_off = i + 1; break; }
    }

    // Extract path
    for (int i = 0; i < 64 && (p_off + i) < len; i++) {
        if (buf[p_off + i] == ' ') break;
        evt->path[i] = buf[p_off + i];
    }
    
    evt->dpi_protocol = 1; // HTTP
}

static __always_inline void parse_dns(struct __sk_buff *ctx, void *data, void *payload,
                                      void *data_end, struct flow_event_t *evt) {
    // DNS header is 12 bytes, question starts after
    void *q = payload + 12;
    if (q >= data_end) return;

    char name[64] = {};
    int len = load_payload(ctx, data, q, data_end, name, sizeof(name));
    

    // Copy DNS query name
    for (int i = 0; i < 64 && i < len; i++) {
        char c = name[i];
        evt->query_name[i] = c;
        if (c == 0) { len = i + 1; break; }
    }

    // Extract query type if available
    if ((void *)q + len + 2 <= data_end) {
        __u16 qtype = 0;
        bpf_skb_load_bytes(ctx, (long)((void *)q - data) + len, &qtype, sizeof(qtype));
        evt->query_type = bpf_ntohs(qtype);
    }

    evt->dpi_protocol = 2; // DNS
}

static __always_inline void parse_icmp(struct icmphdr *icmp, struct flow_event_t *evt) {
    evt->icmp_type = icmp->type;
    evt->dpi_protocol = 3; // ICMP
}

static __always_inline int emit_and_return(struct flow_event_t *evt) {
    // bpf_printk("TC: Submitting packet event, proto=%d\n", evt->protocol);

    if (match_rule(evt)) {
        bpf_printk("TC: Matched rule for drop, discarding event and packet.\n");
        bpf_ringbuf_discard(evt, 0); 
        return TC_ACT_SHOT;
    }

    bpf_ringbuf_submit(evt, 0);
    return TC_ACT_OK;
}

static __always_inline int discard_and_return(struct flow_event_t *evt) {
    bpf_printk("TC: Submitting error event, proto=%d\n", evt->protocol);
    bpf_ringbuf_discard(evt, 0);
    return TC_ACT_OK;
}

static __always_inline int parse_packet(struct __sk_buff *ctx, __u8 direction) {
    void *data     = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;

    struct ethhdr *eth = data;
    if (check_bounds(eth, data_end, sizeof(*eth)))
        return TC_ACT_OK;

    if (bpf_ntohs(eth->h_proto) != ETH_P_IP)
        return TC_ACT_OK;

    struct iphdr *ip = (void *)eth + sizeof(*eth);
    if (check_bounds(ip, data_end, sizeof(*ip)))
        return TC_ACT_OK;

    if (ip->ihl < 5)
        return TC_ACT_OK;

    if ((void *)ip + (ip->ihl * 4) > data_end)
        return TC_ACT_OK;

    if (!(ip->protocol == TCP || ip->protocol == UDP || ip->protocol == ICMP))
        return TC_ACT_OK;

    struct flow_event_t *evt = bpf_ringbuf_reserve(&events, sizeof(*evt), 0);
    if (!evt)
        return TC_ACT_OK;

    void *l4 = (void *)ip + (ip->ihl * 4);

    __builtin_memset(evt, 0, sizeof(*evt));
    evt->timestamp = bpf_ktime_get_ns();
    evt->payload_len = ctx->len;
    evt->direction = direction;
    evt->ifindex = ctx->ifindex;
    evt->src_ip = ip->saddr;
    evt->dst_ip = ip->daddr;
    evt->protocol = ip->protocol;

    if (ip->protocol == TCP) {
        struct tcphdr *tcp = l4;
        if (check_bounds(tcp, data_end, sizeof(*tcp)))
            return discard_and_return(evt);

        if (tcp->doff < 5 || (void *)tcp + (tcp->doff * 4) > data_end)
            return discard_and_return(evt);

        evt->src_port = bpf_ntohs(tcp->source);
        evt->dst_port = bpf_ntohs(tcp->dest);

        if (evt->dst_port == 80 || evt->src_port == 80 ||
            evt->dst_port == 8080 || evt->src_port == 8080) {
            void *payload = (void *)tcp + (tcp->doff * 4);
            parse_http(ctx, data, payload, data_end, evt);
        }

        return emit_and_return(evt);

    } else if (ip->protocol == UDP) {
        struct udphdr *udp = l4;
        if (check_bounds(udp, data_end, sizeof(*udp)))
            return discard_and_return(evt);

        evt->src_port = bpf_ntohs(udp->source);
        evt->dst_port = bpf_ntohs(udp->dest);

        if (evt->dst_port == 53 || evt->src_port == 53) {
            void *payload = (void *)udp + sizeof(*udp);
            parse_dns(ctx, data, payload, data_end, evt);
        }

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
    
// TC ingress program
SEC("tc")
int tc_ingress(struct __sk_buff *ctx) {
    return parse_packet(ctx, 1);
}

// TC egress program
SEC("tc")
int tc_egress(struct __sk_buff *ctx) {
    return parse_packet(ctx, 0);
}

char LICENSE[] SEC("license") = "GPL";