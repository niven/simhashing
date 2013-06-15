package simhashing

// Returns the number of bits set in x
func BitsSet(x uint64) (count int) {

	count = 0
	for ; x > 0; count++ {
		x &= x - 1 // clear the least significant bit set
	}

	return count
}

// split in string into tokens of length
func Tokenize(in string, length int) (tokens []string) {

	// we'll produce ceil(size/len) tokens
	num_tokens := (len(in)-1)/length + 1
	tokens = make([]string, num_tokens)

	i := 0
	for ; i < num_tokens-1; i++ {
		tokens[i] = in[i*length : (i+1)*length]
	}
	tokens[i] = in[i*length:] // last token is just the rest

	return tokens
}

// make tokens in overlaps: 'abcdef',3 -> 'abc','bcd','cde','def'
func Tokenize_stride(in string, length int) (tokens []string) {

	// being paranoid
	if length > len(in) {
		return []string{in}
	}

	num_tokens := len(in) - length + 1
	tokens = make([]string, num_tokens)

	for i := 0; i < num_tokens; i++ {
		tokens[i] = in[i : i+length]
	}

	return tokens
}

// Calculate the hamming distance of 2 64 bit integers
// Defined as: number of bit flips required to turn a into b
func HammingDistance(a uint64, b uint64) int {

	// keep only the bits that are different
	x := a ^ b

	count := 0
	for ; x > 0; count++ {
		x &= x - 1 // clear the least significant bit set
	}

	return count
}
