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

type version_t uint8

const (
	sortedList = version_t(iota)
	unsortedList
)

// DeltaCompPostings represents the provided postings list (which is
// an array of uint64s) in a highly compressed form by calculating
// the deltas and the min number of bits needed to store each delta.
type DeltaCompPostings struct {
	// Metadata
	firstEntry      uint64 // First entry of the provided list
	numPostings     uint32 // Number of entries in the provided list
	numBitsPerDelta uint8  // Min bits needed for storing any delta

	version version_t // Type of postings => sorted, unsorted

	// Data
	data []byte // Bit packed deltas
}

// Returns an empty DeltaCompPostings.
func NewDeltaCompPostings() *DeltaCompPostings {
	return &DeltaCompPostings{}
}

// EncodeSorted derives the deltas from the provided postings list
// and builds the metadata and the minimalist byte array needed for
// storing all the meaningful content of the positings list.
//
// The pre-requisite here is that the provided list needs to be
// sorted.
func (dcp *DeltaCompPostings) EncodeSorted(postings []uint64) error {
	if len(postings) == 0 {
		return fmt.Errorf("EncodeSorted: Empty postings list")
	}

	// Determine the deltas, note that since the first entry in the
	// provided postings list will be saved within metadata, the
	// size of the delta array is one less than the provided list.
	firstEntry := postings[0]
	deltaArray := make([]uint64, len(postings)-1)
	largestDelta := uint64(0)
	for i := 0; i < len(postings)-1; i++ {
		delta := postings[i+1] - postings[i]
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

	// This will be the cursor for the data array.
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
		// For example, consider splitting into bytes 010010100010:
		//     the 2 bytes should be: 01001010 and 00100000
		//     and not: 01001010 and 00000010
		suffix := strings.Repeat("0", 8-len(entry))
		entry = entry + suffix

		x, _ := strconv.ParseUint(entry, 2, 8)
		data[cursor] = byte(x)
	}

	dcp.firstEntry = firstEntry
	dcp.numPostings = uint32(len(postings))
	dcp.numBitsPerDelta = numBitsPerDelta
	dcp.version = sortedList
	dcp.data = data

	return nil
}

// Encode derives the deltas from the provided postings list and
// builds the metadata and the minimalist byte array needed for
// storing all the meaningful content of the positings list.
//
// The provided list CAN be unsorted.
func (dcp *DeltaCompPostings) Encode(postings []uint64) error {
	if len(postings) == 0 {
		return fmt.Errorf("Encode: Empty postings list")
	}

	// Determine the deltas, note that since the first entry in the
	// provided postings list will be saved within metadata, the
	// size of the delta array is one less than the provided list.
	firstEntry := postings[0]
	deltaArray := make([]int64, len(postings)-1)
	largestDelta := uint64(0)
	for i := 0; i < len(postings)-1; i++ {
		delta := int64(postings[i+1] - postings[i])
		delta_edit := delta
		if delta_edit < 0 {
			delta_edit = delta_edit * (-1)
		}
		if uint64(delta_edit) > largestDelta {
			largestDelta = uint64(delta_edit)
		}

		deltaArray[i] = delta
	}

	// Calculate minimum number of bits needed to store every delta,
	// and one extra bit to represent the sign of the delta.
	// For example, if the largest delta without the sign takes 4 bits,
	// 5 bits will be used for the deltas - 1 for the sign, and 4 for value.
	//     A 2 is represented as: 00010.
	//     A -2 is represented asL 10010.
	numBitsPerDelta := uint8(1) + uint8(math.Log2(float64(largestDelta)) + 1)

	// Total bytes needed to hold all the deltas.
	bytesNeededForDeltas :=
		int(math.Ceil(float64(len(deltaArray)*int(numBitsPerDelta)) / 8.0))

	data := make([]byte, bytesNeededForDeltas)

	// This will be the cursor for the data array.
	cursor := 0

	// This will be used to track each byte array to the data array,
	// this string will be populated until it reaches a length of 8.
	// This is so to map it to a single byte later.
	var entry string

	// Iterate over the deltas and split them up in groups of 8 (a byte)
	// and populate the data byte array with them.
	for _, k := range deltaArray {
		k_edit := k
		if k_edit < 0 {
			k_edit = k_edit * (-1)
		}

		delta := strconv.FormatInt(k_edit, 2)
		prefix := strings.Repeat("0", int(numBitsPerDelta)-len(delta))
		if k < 0 {
			// Prefix's length is at least 1, always. So need for any
			// safety check here.
			prefix = "1" + prefix[1:]
		}
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
		// Suffix by zeros so tht the bits are sequentially mapped.
		// For example, consider splitting into bytes 010010100010:
		//     the 2 bytes should be: 01001010 and 00100000
		//     and not: 01001010 and 00000010.
		suffix := strings.Repeat("0", 8-len(entry))
		entry = entry + suffix

		x, _ := strconv.ParseUint(entry, 2, 8)
		data[cursor] = byte(x)
	}

	dcp.firstEntry = firstEntry
	dcp.numPostings = uint32(len(postings))
	dcp.numBitsPerDelta = numBitsPerDelta
	dcp.version = unsortedList
	dcp.data = data

	return nil
}

// Decode decodes the stored delta postings, and returns the
// original postings list.
func (dcp *DeltaCompPostings) Decode() []uint64 {
	if dcp.numPostings == 0 {
		return []uint64{}
	}

	if dcp.version == sortedList {
		return dcp.decodeSorted()
	} else { // unsortedList
		return dcp.decodeUnsorted()
	}
}

func (dcp *DeltaCompPostings) decodeSorted() []uint64 {
	postings := make([]uint64, dcp.numPostings)

	var entry uint64
	shiftBy := dcp.numBitsPerDelta
	entriesAdded := 0

	// Decode the encoded deltas and add them into the
	// postings array instantiated previously.
	for i := 0; i < len(dcp.data); i++ {
		for j := uint(0); j < 8; j++ {
			bit := uint64(dcp.data[i] & (128 >> j) >> uint(7-j))
			shiftBy--

			entry += bit << shiftBy

			if shiftBy == 0 {
				shiftBy = dcp.numBitsPerDelta
				// Increment prior to adding the entry as the first
				// posting is not included in the delta postings.
				entriesAdded++
				postings[entriesAdded] = entry
				entry = 0

				if entriesAdded+1 == len(postings) {
					// All entries added.
					break
				}
			}
		}
	}

	// Convert the built deltas into the postings using
	// the stored metadata.
	var prev uint64
	for i := 0; i < len(postings); i++ {
		prev += postings[i]
		postings[i] = dcp.firstEntry + prev
	}

	return postings
}

func (dcp *DeltaCompPostings) decodeUnsorted() []uint64 {
	deltas := make([]int64, dcp.numPostings-1)

	var entry int64
	shiftBy := dcp.numBitsPerDelta
	entriesAdded := 0
	lastEntrySign := 0

	// Decode the encoded deltas, consider the sign and then
	// add them into the postings array instantiated previously.
	for i := 0; i < len(dcp.data); i++ {
		for j := uint(0); j < 8; j++ {
			if shiftBy == dcp.numBitsPerDelta {
				lastEntrySign = int(dcp.data[i] & (128 >> j) >> uint(7-j))
				shiftBy--
				j++
				// If the sign was in the last bit of a byte, break out and
				// continue with the next byte.
				if j > 7 {
					break
				}
			}

			bit := int64(dcp.data[i] & (128 >> j) >> uint(7-j))
			shiftBy--

			entry += bit << shiftBy

			if shiftBy == 0 {
				shiftBy = dcp.numBitsPerDelta

				if lastEntrySign == 1 {
					entry *= -1
				}

				deltas[entriesAdded] = entry
				entriesAdded++
				entry = 0

				if entriesAdded == len(deltas) {
					// All entries added.
					break
				}
			}
		}
	}

	// Convert the built deltas into the postings using
	// the stored metadata.
	postings := make([]uint64, dcp.numPostings)
	postings[0] = dcp.firstEntry
	for i := 1; i < len(postings); i++ {
		postings[i] = uint64(int64(postings[i-1]) + deltas[i-1])
	}

	return postings
}

// SizeInBytes fetches the footprint of the DeltaCompPostings.
func (dcp *DeltaCompPostings) SizeInBytes() int {
	return 8 /* size of firstEntry (uint64) */ +
		4 /* size of numPostings (uint32) */ +
		1 /* size of numBitsPerDelta (uint8) */ +
		1 /* size of version (uint8) */ +
		len(dcp.data)
}
