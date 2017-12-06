# reductor
This library/tool aims at reducing the footprint of a postings list (an array of uint32s).

## how does it work
Reductor assumes that the provided postings list is already sorted. It estimates the difference or the delta between adjacent entries and stores the deltas which are much smaller values than the original entries. These deltas are then bit-packed (after estimating the minimum number of bits needed to store each delta) to form a highly compressed data structure carrying all the necessary information needed to rebuild the original postings list.

## available APIs
As of now, reductor offers 2 functional APIs - one to encode a provided list (Encode), two to decode a provided list (Decode). An additional API to retrieve the size/footprint of the encoded data structure is also available (SizeInBytes).

## example
Consider the following []uint32:

            101, 105, 215, 218, 240, 260, 280, 290, 320, 325, 375, 480, 578, 690, 755

- This array has a footprint of 15 * 4 = 60 Bytes.
- Using reductor, the footprint of the generated data structure is 22 Bytes.
- That's a reduction of **63.33%**.

## future / to-do
- Accommodating numeric pairs as []{uint32, uint16}, the new uint16 or even uint8 is to denote the term frequencies.
- Multiple blocks to accommodate the postings, so we could potentially further reduce the number of bits used for the deltas.
- Supporting merge operations?
- More services (apis) ...
