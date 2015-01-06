from threading import Thread

i = 0

def thread1Func():
    global i
    for k in range(1000000):
        i += 1

def thread2Func():
    global i
    for k in range(1000000):
        i -= 1

def main():
    global i
    thread1 = Thread(target = thread1Func, args = (),)
    thread2 = Thread(target = thread2Func, args = (),)
    thread1.start()
    thread2.start()
    
    thread1.join()
    thread2.join()

    # This also becomes a somewhat random number, since the threads
    # can be interrupted when Python (?) finds it fit.
    print(i)

main()