#ifndef __TRAFFIC_H__
#define __TRAFFIC_H__

#include <linux/types.h>

// ---------------------------
// DPI Substructures
// ---------------------------

struct http_info_t {
    char method[8];         // GET, POST, etc.
    char path[128];         // URI path (Optional: fill if parsed)
    char host[128];         // Host header (Optional)
    char user_agent[128];   // Optional: User-Agent header
};

struct dns_info_t {
    char query_name[128];   // Domain being queried
    __u16 query_type;       // A=1, AAAA=28, TXT=16, etc.
};

struct arp_info_t {
    __u16 opcode;               // 1 = request, 2 = reply
    __u32 sender_ip;            // Sender protocol address (IPv4)
    __u32 target_ip;            // Target protocol address (IPv4)
    __u8  sender_mac[6];        // Sender hardware address (MAC)
    __u8  target_mac[6];        // Target hardware address (MAC)
};

// ---------------------------
// Main Flow Event Structure
// ---------------------------

struct flow_event_t {
    __u64 timestamp;            // Nanoseconds since boot or epoch

    // Networking metadata

    __u32 src_ip;               // Source IP (IPv4)
    __u32 dst_ip;               // Destination IP (IPv4)
    __u16 src_port;             // Source port (0 for non-TCP/UDP)
    __u16 dst_port;             // Destination port (0 for non-TCP/UDP)
    __u8  protocol;             // TCP = 6, UDP = 17, ICMP = 1, etc.
    __u8  direction;            // 0 = ingress, 1 = egress
    __u16 _pad;                 // Padding for alignment

    __u32 payload_len;          // Payload size (excluding headers)

    __u8  dpi_protocol;         // 0 = unknown, 1 = HTTP, 2 = DNS, 3 = TLS, 4 = ICMP, 5 = ARP
    __u8  icmp_type;            // ICMP type (if applicable)
    __u8  icmp_code;            // ICMP code (if applicable)
    __u8  _pad2;                // Padding for alignment

    //if encrypted
    __u16 tls_version;

    // DPI Protocol Info (fill only one of these based on detected protocol)
    struct http_info_t http;
    struct dns_info_t  dns;    
    struct arp_info_t  arp;
};

#endif // __TRAFFIC_H__
