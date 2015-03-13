import socket
import random

SV_IP = "78.91.19.229"
SV_PORT = 20012
Socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
Socket.bind((SV_IP, SV_PORT))

while True:
    Message, Sender = Socket.recvfrom(1024)
    print Message
    Reply = "Hello back!"
    Socket.sendto(Reply, Sender)
Socket.close()
