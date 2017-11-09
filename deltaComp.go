//
// Reductor: Delta compresses provided postings lists
// Author: Abhinav Dangeti
//

package reductor

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// DeltaCompPostings represents the provided postings list (which is
// an array of uint32s) in a highly compressed form by calculating
// the deltas and the min number of bits needed to store each delta.
type DeltaCompPostings struct {
	// Metadata
	firstEntry      uint32 // First entry of the provided list
	numPostings     uint32 // Number of entries in the provided list
	numBitsPerDelta uint8  // Min bits needed for storing any delta

	// Data
	data []byte // Bit packed deltas
}

// Returns an empty DeltaCompPostings.
func NewDeltaCompPostings() *DeltaCompPostings {
	return &DeltaCompPostings{}
}

// AddAll derives the deltas from the provided postings list and
// builds the metadata and the minimalist byte array needed for
// storing all the meaningful content of the positings list.
//
// The pre-requisite here is that the provided list needs to be
// sorted.
func (dcp *DeltaCompPostings) AddAll(arr []uint32) error {
	if len(arr) == 0 {
		return fmt.Errorf("AddAll: Provided array is empty")
	}

	// Determine the deltas, note that that since the first entry
	// in the provided postings list will be saved within metadata,
	// the size of the delta array is one less than the provided list.
	firstEntry := arr[0]
	deltaArray := make([]uint32, len(arr)-1)
	largestDelta := uint32(0)
	for i := 0; i < len(arr)-1; i++ {
		delta := arr[i+1] - arr[i]
		if delta > largestDelta {
			largestDelta = delta
		}

		deltaArray[i] = delta
	}

	// Calculate minimum number of bits needed to store every delta.
	numBitsPerDelta := uint8(math.Log2(float64(largestDelta)) + 1)

	// Total bytes needed to hold all the deltas.
	bytesNeededForDeltas :=
		int(math.Ceil(float64(len(deltaArray)*int(numBitsPerDelta)) / 8.0))

	data := make([]byte, bytesNeededForDeltas)

	// This will be cursor for the data array.
	cursor := 0

	// This will be used to track each byte entry to the data array,
	// this string will be populated until it reaches a length of 8.
	// This is so to map it to a single byte later.
	var entry string

	// Iterate over the deltas and split them up in groups of 8 (a byte)
	// and populate the data byte array with them.
	for _, k := range deltaArray {
		delta := strconv.FormatInt(int64(k), 2)
		prefix := strings.Repeat("0", int(numBitsPerDelta)-len(delta))
		delta = prefix + delta

		for i := 0; i < len(delta); i++ {
			if len(entry) == 8 {
				// Now that entry has a length of 8 (so it can be mapped to
				// a byte), add it into data after converting it into a byte.
				x, _ := strconv.ParseUint(entry, 2, 8)
				data[cursor] = byte(x)
				cursor++

				// Reset the entry string.
				entry = ""
			}

			entry += string(delta[i])
		}
	}

	// Add any remaining bits into the data byte array.
	if len(entry) > 0 {
		// Last case where, the number of bits left is less than 8 =>
		// Suffix by zeros so that the bits are sequentially mapped.
		// For example, Consider splitting into bytes 010010100010:
		//     the 2 bytes should be: 01001010 and 00100000
		//     and not: 01001010 and 00000010
		suffix := strings.Repeat("0", 8-len(entry))
		entry = entry + suffix

		x, _ := strconv.ParseUint(entry, 2, 8)
		data[cursor] = byte(x)
	}

	dcp.firstEntry = firstEntry
	dcp.numPostings = uint32(len(arr))
	dcp.numBitsPerDelta = numBitsPerDelta
	dcp.data = data

	return nil
}

// FetchAll decodes the stored delta postings, and returns the
// original postings list.
func (dcp *DeltaCompPostings) FetchAll() []uint32 {
	if dcp.numPostings == 0 {
		return []uint32{}
	}

	postings := make([]uint32, dcp.numPostings)

	var entry uint32
	shiftBy := dcp.numBitsPerDelta
	entriesAdded := 0

	// Decode the encoded deltas and add them into the
	// postings array instantiated previously.
	for i := 0; i < len(dcp.data); i++ {
		for j := uint(0); j < 8; j++ {
			shiftBy = shiftBy - 1
			bit := uint32(dcp.data[i] & (128 >> j) >> uint(7-j))

			entry += bit << shiftBy

			if shiftBy == 0 {
				shiftBy = dcp.numBitsPerDelta
				// Increment prior to adding the entry as the first
				// posting is not included in the delta postings.
				entriesAdded++
				postings[entriesAdded] = entry
				entry = 0

				if entriesAdded+1 == len(postings) {
					break
				}
			}
		}
	}

	// Convert the built deltas into the postings using
	// the stored metadata.
	var prev uint32
	for i := 0; i < len(postings); i++ {
		prev += postings[i]
		postings[i] = dcp.firstEntry + prev
	}

	return postings
}

// Len fetches the footprint of the DeltaCompPostings.
func (dcp *DeltaCompPostings) Len() int {
	return 4 /* size of firstEntry (uint32) */ +
		4 /* size of numPostings (uint32) */ +
		1 /* size of numBitsPerDelta (uint8) */ +
		len(dcp.data)
}
