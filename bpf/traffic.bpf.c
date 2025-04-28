#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <traffic.h>

#define TCP 6
#define UDP 17
#define ICMP 1
#define ETH_P_IP 0x0800
#define ETH_P_ARP 0x0806

struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 24);
} events SEC(".maps");

static __always_inline int check_bounds(void *ptr, void *data_end, __u64 size) {
    if (ptr + size > data_end)
        return 1;
    return 0;
}

static __always_inline int parse_arp(struct flow_event_t *packet, void *data, void *data_end) {
    struct arp_hdr {
        __be16 ar_hrd;
        __be16 ar_pro;
        __u8 ar_hln;
        __u8 ar_pln;
        __be16 ar_op;
        __u8 ar_sha[6];
        __u32 ar_sip;
        __u8 ar_tha[6];
        __u32 ar_tip;
    } __attribute__((packed));

    struct arp_hdr *arp = data + sizeof(struct ethhdr);
    if (check_bounds(arp, data_end, sizeof(*arp))) {
        return 1;
    }

    packet->arp.sender_ip = arp->ar_sip;
    packet->arp.target_ip = arp->ar_tip;
    packet->dpi_protocol = 5; // ARP

    return 0;
}

static __always_inline int parse_http(void *payload, void *data_end, struct http_info_t *http) {
    char buf[256];
    int to_copy = (data_end - payload) > 256 ? 256 : (data_end - payload);
    if (bpf_probe_read_kernel(buf, to_copy, payload) < 0) {
        return 1;
    }

    // Parse method and path from the request line
    int m_len = 0;
    #pragma unroll
    for (int i = 0; i < 8; i++) {
        if (buf[i] == ' ') break;
        http->method[i] = buf[i];
        m_len++;
    }

    int p_idx = m_len + 1;
    int path_len = 0;
    #pragma unroll
    for (int i = 0; i < 128; i++) {
        if (p_idx + i >= to_copy || buf[p_idx + i] == ' ') break;
        http->path[i] = buf[p_idx + i];
        path_len++;         
    }

    // Search for Host header
    #pragma unroll
    for (int i = 0; i < 256 - 6; i++) {
        if (i + 6 > to_copy) break;
        if (buf[i] == 'H' && buf[i+1] == 'o' && buf[i+2] == 's' && buf[i+3] == 't' && buf[i+4] == ':' && buf[i+5] == ' ') {
            int h_start = i + 6;
            #pragma unroll
            for (int j = 0; j < 128; j++) {
                if (h_start + j >= to_copy || buf[h_start + j] == '\r' || buf[h_start + j] == '\n') break;
                http->host[j] = buf[h_start + j];
            }
            break;
        }
    }
    return 0;
}

static __always_inline int parse_tls_version(void *payload, void *data_end, __u16 *tls_version) {
    if (payload + 9 > data_end) return 1; // Check bounds for handshake and version fields

    __u8 content_type;
    bpf_probe_read_kernel(&content_type, 1, payload);
    if (content_type != 0x16) return 1; // Not handshake

    __u8 handshake_type;
    bpf_probe_read_kernel(&handshake_type, 1, payload + 5);
    if (handshake_type != 0x01) return 1; // Not ClientHello

    __u16 version;
    bpf_probe_read_kernel(&version, sizeof(version), payload + 1);
    *tls_version = bpf_ntohs(version);

    return 0;
}

static __always_inline int parse_dns(void *payload, void *data_end, struct dns_info_t *dns) {
    if (payload + 12 > data_end) return 1;  // DNS header is 12 bytes

    void *q = payload + 12; // Question section starts after the 12-byte header
    #pragma unroll
    for (int i = 0; i < 128; i++) {
        if (q + i + 1 > data_end) break;
        char c;
        bpf_probe_read_kernel(&c, 1, q + i);
        if (c == 0) {
            dns->query_name[i] = 0;
            break;
        }
        dns->query_name[i] = c;
    }
    __u16 *qt = q + __builtin_strlen(dns->query_name) + 1;
    if ((void *)(qt + 1) <= data_end) {
        __u16 qtype;
        bpf_probe_read_kernel(&qtype, sizeof(qtype), qt);
        dns->query_type = bpf_ntohs(qtype);
    }
    return 0;
}

static __always_inline int parse_packet(struct __sk_buff *ctx, __u8 direction) {
    struct flow_event_t *packet = bpf_ringbuf_reserve(&events, sizeof(*packet), 0);
    if (!packet) {
        return 0;
    }

    __builtin_memset(packet, 0, sizeof(*packet));

    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;

    packet->payload_len = ctx->len;
    packet->direction = direction;
    packet->timestamp = bpf_ktime_get_ns();

    struct ethhdr *eth = data;
    if (check_bounds(eth, data_end, sizeof(*eth))) {
        bpf_ringbuf_discard(packet, 0);
        return 0;
    }

    __u16 eth_proto = bpf_ntohs(eth->h_proto);

    if (eth_proto == ETH_P_ARP) {
        if (!parse_arp(packet, data, data_end)) {
            bpf_ringbuf_submit(packet, 0);
        } else {
            bpf_ringbuf_discard(packet, 0);
        }
        return 0;
    }

    if (eth_proto != ETH_P_IP) {
        bpf_ringbuf_discard(packet, 0);
        return 0;
    }

    struct iphdr *ip = (void *)(eth + 1);
    if (check_bounds(ip, data_end, sizeof(*ip))) {
        bpf_ringbuf_discard(packet, 0);
        return 0;
    }

    packet->src_ip = ip->saddr;
    packet->dst_ip = ip->daddr;
    packet->protocol = ip->protocol;

    void *l4 = (void *)ip + (ip->ihl * 4);
    if (l4 > data_end) {
        bpf_ringbuf_discard(packet, 0);
        return 0;
    }

    if (ip->protocol == TCP) {
        struct tcphdr *tcp = l4;
        if (check_bounds(tcp, data_end, sizeof(*tcp))) {
            bpf_ringbuf_discard(packet, 0);
            return 0;
        }
        packet->src_port = bpf_ntohs(tcp->source);
        packet->dst_port = bpf_ntohs(tcp->dest);

        void *payload = (void *)tcp + (tcp->doff * 4);
        if (payload <= data_end) {
            if (packet->dst_port == 80) {
                packet->dpi_protocol = 1; // HTTP
                parse_http(payload, data_end, &packet->http);
            } else if (packet->dst_port == 443) {
                packet->dpi_protocol = 3; // TLS
                parse_tls_version(payload, data_end, &packet->tls_version);
            }
        }

    } else if (ip->protocol == UDP) {
        struct udphdr *udp = l4;
        if (check_bounds(udp, data_end, sizeof(*udp))) {
            bpf_ringbuf_discard(packet, 0);
            return 0;
        }
        packet->src_port = bpf_ntohs(udp->source);
        packet->dst_port = bpf_ntohs(udp->dest);

        void *payload = (void *)udp + sizeof(*udp);
        if (payload <= data_end && (packet->dst_port == 53 || packet->src_port == 53)) {
            packet->dpi_protocol = 2; // DNS
            parse_dns(payload, data_end, &packet->dns);
        }

    } else if (ip->protocol == ICMP) {
        struct icmphdr *icmp = l4;
        if (check_bounds(icmp, data_end, sizeof(*icmp))) {
            bpf_ringbuf_discard(packet, 0);
            return 0;
        }
        packet->dpi_protocol = 4;
        packet->icmp_type = icmp->type;
        packet->icmp_code = icmp->code;

        bpf_ringbuf_submit(packet, 0);
        return 0;
    } else {
        bpf_ringbuf_discard(packet, 0);
        return 0;
    }
}

SEC("cgroup_skb/ingress")
int attach_ingress(struct __sk_buff *ctx) {
    return parse_packet(ctx, 1);
}

SEC("cgroup_skb/egress")
int attach_egress(struct __sk_buff *ctx) {
    return parse_packet(ctx, 0);
}

char LICENSE[] SEC("license") = "GPL";