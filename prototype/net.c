#include "net.h"
#include <netdb.h>
#include <unistd.h>
#include <sys/socket.h>

#define RECV_SIZE 4096
static int g_socket;
// static int g_preferred_listen_port;

int
net_initialize(uint16 listen_port)
{
    g_socket = socket(AF_INET, SOCK_DGRAM, 0);
    if (g_socket < 0)
    {
        printf("Failed to create a socket\n");
        return FAIL;
    }

    struct sockaddr_in address = {};
    address.sin_family = AF_INET;
    address.sin_addr.s_addr = INADDR_ANY;
    address.sin_port = htons(listen_port);

    if (bind(g_socket, (struct sockaddr *)&address, sizeof(address)) < 0)
    {
        printf("Failed to bind socket\n");
        return FAIL;
    }

    return OK;
}

int
net_send(struct NetAddress *destination, char *data, int length)
{
    struct sockaddr_in address;
    address.sin_family = AF_INET;
    address.sin_addr.s_addr = htonl(
        (destination->ip0 << 24) |
        (destination->ip1 << 16) |
        (destination->ip2 <<  8) |
        (destination->ip3));
    address.sin_port = htons(destination->port);

    return sendto(g_socket, data, length, 0,
           (struct sockaddr*)&address, sizeof(struct sockaddr_in));
}

int
net_read(char *data, int max_size, struct NetAddress *sender)
{
    struct sockaddr_in from;
    int from_length = sizeof(from);
    int bytes_read = recvfrom(
        g_socket, data, max_size, 0,
        (struct sockaddr*)&from, &from_length);
    if (bytes_read <= 0)
        return 0;

    uint32 from_address = ntohl(from.sin_addr.s_addr);
    sender->ip0  = (from_address >> 24) & 0xff;
    sender->ip1  = (from_address >> 16) & 0xff;
    sender->ip2  = (from_address >>  8) & 0xff;
    sender->ip3  = (from_address >>  0) & 0xff;
    sender->port = ntohs(from.sin_port);

    printf("Read %d bytes from %d.%d.%d.%d:%d\n",
            bytes_read, sender->ip0, sender->ip1,
            sender->ip2, sender->ip3, sender->port);

    return bytes_read;
}