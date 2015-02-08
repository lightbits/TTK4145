#include "net.h"
#include "net.c"

struct Message
{
    uint32 protocol;
    uint16 flag;
    char hail[256];
};

int
main(int argc, char **argv)
{
    if (net_initialize(43002) != OK)
    {
        printf("Failed to initialize socket\n");
        return FAIL;
    }

    printf("hi!");

    struct NetAddress sender = {};
    char data[256];
    int bytes_read = net_read(data, sizeof(data), &sender);

    struct Message msg = {};
    msg.protocol = 0xabad1dea;
    msg.flag = 0xbeef;
    sprintf(msg.hail, "Client says hi!");
    int bytes_sent = net_send(&sender, (char*)&msg, sizeof(struct Message));
    
    return OK;
}