#include "net.h"
#include <netdb.h>
#include <unistd.h>
#include <sys/socket.h>

#define RECV_SIZE 4096
#define DEFAULT_LISTEN_PORT 20012
typedef struct
{
    int socket;
    bool initialized;
} Network;

global_variable Network network;

bool
net_init(uint16 listen_port)
{
    network.initialized = false;
    network.socket = socket(AF_INET, SOCK_DGRAM, 0);
    if (network.socket < 0)
    {
        printf("Failed to create a socket\n");
        return FAIL;
    }

    struct sockaddr_in address = {};
    address.sin_family = AF_INET;
    address.sin_addr.s_addr = INADDR_ANY;
    address.sin_port = htons(listen_port);

    if (bind(network.socket, (struct sockaddr *)&address, sizeof(address)) < 0)
    {
        printf("Failed to bind socket\n");
        return FAIL;
    }

    network.initialized = true;

    return OK;
}

int
net_send(NetAddress *destination, char *data, int length)
{
    if (!network.initialized)
    {
        printf("Warning: Using default listen port.\n");
        if (!net_init(DEFAULT_LISTEN_PORT))
            return 0;
    }

    struct sockaddr_in address;
    address.sin_family = AF_INET;
    address.sin_addr.s_addr = htonl(
        (destination->ip0 << 24) |
        (destination->ip1 << 16) |
        (destination->ip2 <<  8) |
        (destination->ip3));
    address.sin_port = htons(destination->port);

    int bytes_sent = sendto(network.socket, data, length, 0,
        (struct sockaddr*)&address, sizeof(struct sockaddr_in));
    
    printf("Sent %d bytes to %d.%d.%d.%d:%d\n",
            bytes_sent, destination->ip0, destination->ip1,
            destination->ip2, destination->ip3, destination->port);

    return bytes_sent;
}

int
net_read(char *data, int max_size, NetAddress *sender)
{
    if (!network.initialized)
    {
        printf("Warning: Using default listen port.\n");
        if (!net_init(DEFAULT_LISTEN_PORT))
            return 0;
    }

    struct sockaddr_in from;
    int from_length = sizeof(from);
    int bytes_read = recvfrom(
        network.socket, data, max_size, 0,
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