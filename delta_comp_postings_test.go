//
// Tests for delta compressed postings lists
// Author: Abhinav Dangeti
//

package main

import (
	"fmt"
	"testing"
)

func checkEq(a, b []uint32) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func testBasic(t *testing.T, postings []uint32) {
	dcp := NewDeltaCompPostings()
	dcp.AddAll(postings)

	got := dcp.FetchAll()
	if !checkEq(postings, got) {
		t.Errorf("Expected: %v, Got: %v", postings, got)
	}

	fmt.Printf("Achieved a compression from %v bytes to %v bytes => %.4v%%\n",
		len(postings)*4, dcp.Len(),
		float64(len(postings)*4-dcp.Len())*100/float64(len(postings)*4))

}

func TestBasic1(t *testing.T) {
	postings := []uint32{100, 102, 104, 108, 110}
	testBasic(t, postings)
}

func TestBasic2(t *testing.T) {
	postings := []uint32{
		100, 102, 104, 108, 110,
		120, 140, 200, 500, 622,
		1402, 1550, 2000, 2529}

	testBasic(t, postings)
}
