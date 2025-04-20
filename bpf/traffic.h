#ifndef __TRAFFIC_H__
#define __TRAFFIC_H__



struct flow_event_t {
    __u32 src_ip; // 4 bytes , 0 -> 4
    __u32 dst_ip; // 4 bytes , 4 -> 8
    __u16 src_port; // 2 bytes , 8 -> 10 
    __u16 dst_port; // 2 bytes , 10 -> 12 
    __u8 protocol; // TCP = 6, UDP = 17 , 1 byte , 12 -> 13 
    // till now 9 bytes , need to start at 12 so c copiler add 3 more bytes to padding ! 
    __u64 pkt_len;
};

#endif
