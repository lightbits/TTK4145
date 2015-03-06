// Wrapper for libComedi I/O.
// These functions provide and interface to libComedi limited to use in
// the real time lab.
//
// 2006, Martin Korsgaard
#ifndef __INCLUDE_IO_H__
#define __INCLUDE_IO_H__
#define ET_simulation 0
#define ET_comedi 1

#define ELEVATOR_TYPE ET_comedi  // Use real lift
// #define ELEVATOR_TYPE ET_simulation // Use simulator

// Return 0 on init failure
int  io_init();

void io_set_bit(int channel);
void io_clear_bit(int channel);
void io_write_analog(int channel, int value);
int  io_read_bit(int channel);
int  io_read_analog(int channel);

#endif
