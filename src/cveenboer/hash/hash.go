package hash

import "cveenboer/util"

// Generate a 64 bit simhash for a string
func SimHash(src string) uint64 {

	// Magic number 3, as well as a tokenize function
	tokens := util.Tokenize_stride(src, 3)

	var counts [64]int

	for _, token := range tokens {
		h := Strong64(token) // Could pick any hashing function

		for i := uint8(0); i < 64; i++ {
			if h&(1<<i) > 0 {
				counts[i]++
			} else {
				counts[i]--
			}
		}

	}

	// simhash bit i is 1 if counts[i] > 0 and 0 otherwise

	var simhash uint64
	for i := uint8(0); i < 64; i++ {
		if counts[i] > 0 {
			simhash |= 1 << i
		}
	}

	return simhash
}

// Calculate a basic hash code for a string
func Basic64(in string) uint64 {

	var hash uint64 = 0

	for i := 0; i < len(in); i++ {
		hash = (hash << 5) - hash + uint64(in[i])
	}

	return hash
}

// Strong 64 bit hash for a string.
// references: http://www.javamex.com/tutorials/collections/strong_hash_code.shtml
func Strong64(in string) uint64 {

	h := _HSTART
	var ch uint8
	num_bytes := len(in)

	for i := 0; i < num_bytes; i++ {
		ch = in[i]
		h = (h * _HMULT) ^ byteTable[ch&0xff]
		ch_ror_64 := (ch >> 8) | (ch << (64*8 - 8))
		h = (h * _HMULT) ^ byteTable[ch_ror_64&0xff]
	}

	return h
}

const _HSTART uint64 = 0xBB40E64DA205B064
const _HMULT uint64 = 7664345821815920749

var byteTable [256]uint64

// Initialize the byte table for the strong hashing function
func init() {

	var h uint64 = 0x544B2FBACAAF1684

	for i := 0; i < 256; i++ {
		for j := 0; j < 31; j++ {
			h = (h >> 10) | (h << (64*8 - 7)) ^ h
			h = (h << 11) ^ h
			h = (h >> 10) | (h << (64*8 - 10)) ^ h
		}
		byteTable[i] = h
	}

}
