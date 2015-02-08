#ifndef _net_h_
#define _net_h_
#include <stdio.h>
#include <stdint.h>
#define uint8  uint8_t
#define uint16 uint16_t
#define uint32 uint32_t
#define FAIL -1
#define OK 0

struct NetAddress
{
    uint8 ip0, ip1, ip2, ip3;
    uint16 port;
};

// TODO: replace initialize with set_preferred_listen_port
// do initialization implicit in send/read
int net_initialize(uint16 listen_port);
int net_send(struct NetAddress *destination, char *data, int length);
int net_read(char *data, int max_size, struct NetAddress *sender);

#endif