#include "net.h"
#include "net.c"
#include <stdlib.h>
#include <string.h>

// typedef enum
// {
//     ORDER_DOWN,
//     ORDER_UP,
//     ORDER_OUT
// } OrderType;

// typedef struct
// {
//     int from;
//     int to;
//     OrderType type;
// } Order;

// typedef enum
// {
//     MSG_NEW_ORDER,
//     MSG_
// } MessageType;

// typedef struct
// {
//     MessageType type;
// } Message;

// void
// net_sendall(char *data, int length)
// {

// }

// void
// got_message_from_net(Message msg)
// {
//     switch (msg.type)
//     {
//         case MSG_NEW_ORDER:
//         break;
//         case MSG_
//     }
// }

typedef struct
{
    uint32 protocol;
    uint16 flag;
    char hail[256];
} Message;

int
main(int argc, char **argv)
{
    int listen_port = atoi(argv[1]);
    if (net_init(listen_port) != OK)
    {
        printf("Failed to initialize socket\n");
        return FAIL;
    }

    if (strcmp(argv[2], "client") == 0)
    {
        printf("Waiting to send!\n");
        NetAddress destination = {127, 0, 0, 1, 15432};
        Message msg = {};
        msg.protocol = 0xabad1dea;
        msg.flag = 0xbeef;
        sprintf(msg.hail, "Client says hi!");
        int bytes_sent = net_send(&destination, (char*)&msg, sizeof(Message));
    }
    else
    {
        printf("Waiting to read!\n");
        NetAddress sender = {};
        char data[256];
        int bytes_read = net_read(data, sizeof(data), &sender);
        Message *msg = (Message*)data;
        printf("%x %x %s\n", msg->protocol, msg->flag, msg->hail);
    }
    
    return OK;
}