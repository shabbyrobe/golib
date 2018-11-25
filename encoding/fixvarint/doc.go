/*
Package fixvarint implements "varint" encoding of 64-bit integers optimised for
fixed precision decimals (i.e. decimals that will regularly have trailing decimal
zeros).

It is very similar to the format used in the golang binary package, with a
significant difference: four bits of the first byte (after the first
continuation bit) are set aside to count the number of trailing zeros. This
means numbers from 8-127 are forced to use 2 bytes instead of 1, but with the
benefit that numbers with a large number of trailing decimal zeros are
significantly smaller.

The encoding is:
- the first byte follows the format:
	  0    continuation bit (1 if there are more bytes)
	  1-4  count number of trailing zeros
	  5-7  the 3 least significant bits
- subsequent bits follow the format:
	  0    continuation bit (1 if there are more bytes)
	  1-7  next least significant 7 bits
- the most significant bit (msb) in each output byte indicates if there
  is a continuation byte (msb = 1)
- signed integers are mapped to unsigned integers using "zig-zag"
  encoding: Positive values x are written as 2*x + 0, negative values
  are written as 2*(^x) + 1; that is, negative numbers are complemented
  and whether to complement is encoded in bit 0.

At most 10 bytes are needed for 64-bit values.
*/

package fixvarint
