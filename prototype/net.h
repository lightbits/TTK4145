#ifndef _net_h_
#define _net_h_
#include <stdio.h>
#include <stdint.h>
#define global_variable static
#define uint8  uint8_t
#define uint16 uint16_t
#define uint32 uint32_t
#define int8   int8_t
#define int16  int16_t
#define int32  int32_t
#define bool   int32_t
#define FAIL   0
#define OK     1
#define false  0
#define true   1

struct NetAddress
{
    uint8 ip0, ip1, ip2, ip3;
    uint16 port;
};

bool net_init(uint16 listen_port);
int  net_send(struct NetAddress *destination, char *data, int length);
int  net_read(char *data, int max_size, struct NetAddress *sender);

#endif