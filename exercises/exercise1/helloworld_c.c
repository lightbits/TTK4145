// gcc 4.7.2 +
// gcc -std=gnu99 -Wall -g -o helloworld_c helloworld_c.c -lpthread

#include <pthread.h>
#include <stdio.h>

int i = 0;

// Note the return type: void*
void* thread1Func() {
    for (int k = 0; k < 1000000; k++)
        i++;
    return NULL;
}

void* thread2Func() {
    for (int k = 0; k < 1000000; k++)
        i--;
    return NULL;
}

int main() {
    pthread_t thread1, thread2;
    pthread_create(&thread1, NULL, thread1Func, NULL);
    pthread_create(&thread2, NULL, thread1Func, NULL);
    
    pthread_join(thread1, NULL);
    pthread_join(thread2, NULL);

    // This is more interesting. These are native threads, and as such they
    // can be interrupted down on the lowest level. For instance, in executing
    // thread1, the thread can be interrupted right after fetching the current value
    // of i from memory, then thread2 decrements i, and thread1 finishes
    // by incrementing i and storing in back in memory. But thread1 worked on an
    // old value, instead of the new value that thread2 computed.

    // Example:
    // i = 0
    // thread1 fetches i (i = 0)
    // thread2 interrupts thread1
    // thread2 fetches i (i = 0)
    // thread2 i-- (i = 0 - 1)
    // thread2 write i (i = -1)
    // thread1 i++ (i = 0 + 1)
    // thread1 write (i = +1) !!

    printf("%d\n", i);

    // On my machine I often get values above 1 000 000 when running this program.
    return 0;
}
