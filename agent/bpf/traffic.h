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
    char method[8];     // HTTP method (GET, POST, etc.)
    char path[64];      // Request path
    char query_name[64];  // DNS query name
    __u16 query_type;     // e.g., A=1, AAAA=28
    __u8 icmp_type; 
    __u32 ifindex   ;
    
};


struct flow_rule_t {
    __u32 src_ip;        // 4 bytes
    __u32 dst_ip;        // 4 bytes  
    __u16 src_port;      // 2 bytes
    __u16 dst_port;      // 2 bytes
    __u8 protocol;       // 1 byte
    __u8 direction;      // 1 byte
    __u8 dpi_protocol;   // 1 byte
    __u8 action;         // 1 byte
    __u8 method[8];      // 8 bytes
    __u8 path[64];       // 64 bytes
    __u8 query_name[64]; // 64 bytes
    __u16 query_type;    // 2 bytes
    __u8 icmp_type;      // 1 byte
    // Total: 155 bytes (but compiler will pad to 156 for alignment)
};





#endif // __TRAFFIC_H__
