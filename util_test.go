package simhashing

import "testing"

func TestBitsSet(t *testing.T) {

	if BitsSet(0x1) != 1 {
		t.Errorf("Incorrect count")
	}
	if BitsSet(0x1<<32) != 1 {
		t.Errorf("Incorrect count")
	}
	if BitsSet(0xff) != 8 {
		t.Errorf("Incorrect count")
	}
	if BitsSet(0xf0) != 4 {
		t.Errorf("Incorrect count")
	}

}

func TestTokenize(t *testing.T) {

	if !stringarray_equal(Tokenize("abcdef", 3), []string{"abc", "def"}) {
		t.Errorf("Tokening fail")
	}

	if !stringarray_equal(Tokenize("abcdef", 4), []string{"abcd", "ef"}) {
		t.Errorf("Tokening fail")
	}

	if !stringarray_equal(Tokenize("abcdef", 10), []string{"abcdef"}) {
		t.Errorf("Tokening fail")
	}

}

func TestTokenize_stride(t *testing.T) {

	if !stringarray_equal(Tokenize_stride("abcdef", 3), []string{"abc", "bcd", "cde", "def"}) {
		t.Errorf("Tokening/stride fail")
	}

	if !stringarray_equal(Tokenize_stride("abcdef", 4), []string{"abcd", "bcde", "cdef"}) {
		t.Errorf("Tokening/stride fail")
	}

	if !stringarray_equal(Tokenize_stride("abcdef", 10), []string{"abcdef"}) {
		t.Errorf("Tokening/stride fail")
	}

}

// check that 2 arrays of strings are equal (in contents and order)
func stringarray_equal(a []string, b []string) (equal bool) {

	equal = true

	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return
}
