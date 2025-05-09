#ifndef __TRAFFIC_H__
#define __TRAFFIC_H__

// ---------------------------
// Main Flow Event Structure
// ---------------------------

struct flow_event_t {
    __u64 timestamp;            // Nanoseconds since boot or epoch

    __u32 src_ip;               // Source IP (IPv4)
    __u32 dst_ip;               // Destination IP (IPv4)
    __u16 src_port;             // Source port
    __u16 dst_port;             // Destination port
    __u8  protocol;             // TCP=6, UDP=17, ICMP=1, etc.
    __u8  direction;            // 0 = ingress, 1 = egress
    __u16 payload_len;          // Payload size (bytes, excluding headers)

    __u8  dpi_protocol;         // 0=unknown, 1=HTTP, 2=DNS, 3=ICMP
    __u8  reserved1;            // Alignment padding
    __u16 reserved2;            // Alignment padding

    union {
        struct {
            char method[8];     // HTTP method (GET, POST, etc.)
            char path[64];      // Request path
        } http;

        struct {
            char query_name[64];  // DNS query name
            __u16 query_type;     // e.g., A=1, AAAA=28
        } dns;

        struct {
            __u8 icmp_type;     // ICMP type (e.g. Echo, Time Exceeded)
        } icmp;
    };

    __u32 Pid ; 
};

#endif // __TRAFFIC_H__
