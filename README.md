[![Go Report Card](https://goreportcard.com/badge/abhinavdangeti/reductor)](https://goreportcard.com/report/abhinavdangeti/reductor)
[![Build Status](https://travis-ci.org/abhinavdangeti/reductor.svg?branch=master)](https://travis-ci.org/abhinavdangeti/reductor)
[![Coverage Status](https://coveralls.io/repos/github/abhinavdangeti/reductor/badge.svg?branch=master)](https://coveralls.io/github/abhinavdangeti/reductor?branch=master)
[![GoDoc](https://godoc.org/github.com/abhinavdangeti/reductor?status.svg)](https://godoc.org/github.com/abhinavdangeti/reductor)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# reductor
This library/tool aims at reducing the footprint of a postings list (an array of uint64s).

## how does it work
Reductor can work with sorted and unsorted postings lists. It first estimates the difference or the deltas between adjacent entries. These deltas are typically much smaller values than the original entries. These deltas are then bit-packed (after estimating the minimum number of bits needed to store each delta) to form a highly compressed data structure carrying all the necessary information needed to rebuild the original postings list.

## available APIs
Reductor offers the following APIs:

- EncodeSorted (Encodes ONLY a sorted []uint64)
- Encode (Encodes a []uint64 - sorted/unsorted)
- Decode (Decodes a previously encoded postings list)
- SizeInBytes (Fetches the size in bytes of the encoded postings list)

## example
Consider the following []uint64:

            101, 105, 215, 218, 240, 260, 280, 290, 320, 325, 375, 480, 578, 690, 755

- This array has a footprint of 15 * 8 + 24 (overhead) = 144 Bytes.
- Taking advantage of the fact that the list is sorted, using reductor's EncodeSorted(), the footprint of the generated data structure is 53 Bytes.
- That's a reduction of **63.19%**.

If the same list were presented unsorted:

            280, 105, 215, 690, 240, 578, 101, 320, 755, 325, 375, 480, 260, 218, 290

- Using reductor's Encode(), the footprint of the generated data structure is 58 Bytes.
- That's a reduction of **59.72%**.

## future work
- Split the encoding into multiple blocks of postings. This would further reduce the number of bits used for the deltas, and most importantly - quicken point lookup.
- APIs to support operations over multiple postings lists:
    - Merge
    - Intersection/Union/Difference
