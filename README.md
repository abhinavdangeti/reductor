[![Go Report Card](https://goreportcard.com/badge/abhinavdangeti/reductor)](https://goreportcard.com/report/abhinavdangeti/reductor)
[![Build Status](https://travis-ci.org/abhinavdangeti/reductor.svg?branch=master)](https://travis-ci.org/abhinavdangeti/reductor)
[![GoDoc](https://godoc.org/github.com/abhinavdangeti/reductor?status.svg)](https://godoc.org/github.com/abhinavdangeti/reductor)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# reductor
This library/tool aims at reducing the footprint of a postings list (an array of uint64s).

## how does it work
Reductor assumes that the provided postings list is already sorted. It estimates the difference or the delta between adjacent entries and stores the deltas which are much smaller values than the original entries. These deltas are then bit-packed (after estimating the minimum number of bits needed to store each delta) to form a highly compressed data structure carrying all the necessary information needed to rebuild the original postings list.

## available APIs
Reductor offers the following APIs:

- EncodeSorted (This is to encode a sorted []uint64)
- Encode (This is to encode a []uint64 that could be unsorted)
- Decode (This is to decode a previously encoded postings list - be it sorted or not)
- SizeInBytes (This is to fetch the size in bytes of the encoded postings list)

## example
Consider the following []uint64:

            101, 105, 215, 218, 240, 260, 280, 290, 320, 325, 375, 480, 578, 690, 755

- This array has a footprint of 15 * 8 = 120 Bytes.
- Using reductor, and taking advantage of the fact that the list is sorted, the footprint of the generated data structure is 27 Bytes.
- That's a reduction of **77.5%**.

If the same list were presented unsorted:

            280, 105, 215, 690, 240, 578, 101, 320, 755, 325, 375, 480, 260, 218, 290

- Using reductor, the footprint of the generated data structure is 32 Bytes.
- That's a reduction of **73.33%**.

## future / to-do
- Multiple blocks to accommodate the postings, so we could potentially further reduce the number of bits used for the deltas, and quicken lookup.
- Supporting merge operations
- More services (apis) ...
