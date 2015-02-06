#include <stdio.h>
#include <netdb.h>
#include <unistd.h>

int
main(int argc, char **argv)
{
    int g_socket = socket(AF_INET, SOCK_DGRAM, IPPROTO_UDP);
    if (g_socket <= 0)
    {
        printf("Failed to create a socket\n");
        return -1;
    }

    // Bind socket to a port
    struct sockaddr_in address;
    address.sin_family = AF_INET;
    address.sin_addr.s_addr = INADDR_ANY;
    address.sin_port = htons(20012);

    if (bind(g_socket, (const struct sockaddr*)&address, sizeof(sockaddr_in)) < 0)
    {
        printf("Failed to bind socket\n");
        return -1;
    }

    printf("hi!");
}