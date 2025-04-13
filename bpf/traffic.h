#ifndef __TRAFFIC_H__
#define __TRAFFIC_H__



struct flow_event_t {
    __u32 src_ip;
    __u32 dst_ip;
    __u16 src_port;
    __u16 dst_port;
    __u8 protocol; // TCP = 6, UDP = 17
    __u64 pkt_len;
};

#endif
