// gcc 4.7.2 +
// gcc -std=gnu99 -Wall -g -o helloworld_c helloworld_c.c -lpthread

#include <pthread.h>
#include <stdio.h>

pthread_mutex_t lock_i;
int i = 0;

// Note the return type: void*
void* thread1Func() {
    pthread_mutex_lock(&lock_i);
    for (int k = 0; k < 1000000; k++)
    {
        i++;
    }
    pthread_mutex_unlock(&lock_i);
    return NULL;
}

void* thread2Func() {
    pthread_mutex_lock(&lock_i);
    for (int k = 0; k < 1000001; k++)
    {
        i--;
    }
    pthread_mutex_unlock(&lock_i);
    return NULL;
}

int main() {
    pthread_mutex_init(&lock_i, NULL);

    pthread_t thread1, thread2;
    pthread_create(&thread1, NULL, thread1Func, NULL);
    pthread_create(&thread2, NULL, thread2Func, NULL);
    
    pthread_join(thread1, NULL);
    pthread_join(thread2, NULL);

    printf("%d\n", i);

    pthread_mutex_destroy(&lock_i);
    return 0;
}
