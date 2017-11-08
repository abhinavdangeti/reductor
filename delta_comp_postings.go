//
// Delta compresses provided postings lists
// Author: Abhinav Dangeti
//

package main

import (
	"math"
	"strconv"
)

// Single block implementation
type DeltaCompPostings struct {
	// Metadata
	startDocID      uint32
	numPostings     uint32
	numBitsPerDelta uint8

	// Data
	postingDeltas []byte
}

func NewDeltaCompPostings() *DeltaCompPostings {
	return &DeltaCompPostings{}
}

func (dcp *DeltaCompPostings) AddAll(arr []uint32) {
	if len(arr) == 0 {
		return
	}

	// Assumes arr to be a sorted one
	startDocID := arr[0]
	zeroIndexedArray := make([]uint32, len(arr)-1)
	largestDelta := uint32(0)
	for i := 0; i < len(arr)-1; i++ {
		delta := arr[i+1] - arr[i]
		if delta > largestDelta {
			largestDelta = delta
		}

		zeroIndexedArray[i] = delta
	}

	// Calculate minimum number of bits needed to store any delta
	numBitsPerDelta := uint8(math.Log2(float64(largestDelta)) + 1)

	dcp.startDocID = startDocID
	dcp.numPostings = uint32(len(arr))
	dcp.numBitsPerDelta = numBitsPerDelta

	bytesNeededForDeltas := int(math.Ceil(float64(len(zeroIndexedArray)*int(numBitsPerDelta)) / 8.0))

	postingDeltas := make([]byte, bytesNeededForDeltas)

	var bits string
	for _, entry := range zeroIndexedArray {
		i := strconv.FormatInt(int64(entry), 2)
		j := ""
		if len(i) < int(numBitsPerDelta) {
			for k := 0; k < int(numBitsPerDelta)-len(i); k++ {
				j = j + "0"
			}
		}
		bits = bits + j + i
	}

	index := 0
	for i := 0; i < len(bits); i += 8 {
		j := 8 + i
		if j >= len(bits) {
			j = len(bits)
		}

		x, _ := strconv.ParseInt(bits[i:j], 2, 8)
		// Left shift by j - i if j - i < 8 so that bits are sequentially mapped
		// For example,
		// for the case 010010100010:
		// the 2 bytes should be: 01001010 and 00100000
		// and not: 01001010 and 00000010
		if j-i < 8 {
			x = x << uint(j-i)
		}

		postingDeltas[index] = byte(x)
		index++
	}

	dcp.postingDeltas = postingDeltas
}

func (dcp *DeltaCompPostings) FetchAll() []uint32 {
	postings := make([]uint32, dcp.numPostings)

	var entry uint32
	shiftBy := dcp.numBitsPerDelta
	entriesAdded := 0

	for i := 0; i < len(dcp.postingDeltas); i++ {
		for j := uint(0); j < 8; j++ {
			shiftBy = shiftBy - 1
			bit := uint32(dcp.postingDeltas[i] & (128 >> j) >> uint(7-j))

			if shiftBy == 0 {
				shiftBy = dcp.numBitsPerDelta
				// Increment prior to adding the entry as the first posting
				// is not included in the postings delta list
				entriesAdded++
				postings[entriesAdded] = entry
				entry = 0

				if entriesAdded+1 == len(postings) {
					break
				}
			}

			entry += bit << shiftBy
		}
	}

	var prev uint32
	for i := 0; i < len(postings); i++ {
		prev += postings[i]
		postings[i] = dcp.startDocID + prev
	}

	return postings
}

func (dcp *DeltaCompPostings) Len() int {
	return 4 /* size of startDocID (uint32) */ +
		4 /* size of numPostings (uint32) */ +
		1 /* size of numBitsPerDelta (uint8) */ +
		len(dcp.postingDeltas)
}
