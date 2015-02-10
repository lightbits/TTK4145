// NetSetPreferredListenPort(12345);

//     if (SDL_Init(SDL_INIT_EVERYTHING) != 0)
//         return -1;

//     while (true) 
//     {
//         NetAddress destination = {129, 241, 187, 136, 43002};
//         const char *data = "Hello to you!";
//         int bytes_sent = NetSend(&destination, data, 14);
//         printf("Sent %d bytes\n", bytes_sent);
//         SDL_Delay(1000);

//         NetAddress sender = {};
//         Message reply = {};
//         int bytes_read = NetRead((char*)&reply, sizeof(Message), &sender);
//         printf("Read %d bytes: %x %x %s\n", 
//                bytes_read, reply.protocol, reply.flag, reply.hail);
//     }
//     NetClose();