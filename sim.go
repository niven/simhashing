// This implements a datastore for simhashes (http://matpalm.com/resemblance/simhash/)
// that allows for (at one point) efficient searching of items within some Hamming Distance of a know text
// Currently we do reasonable as long as we have stored many items (~20M) and we search within a range of 8 or so.
package simhashing

import "fmt"
import "container/heap"

const bit_length = 8
const size = 1 << bit_length

const max_keys_per_node = 256
const bits_per_key = 8

// both of these assume bits_per_key equals 8
var level_chunks = [...]uint64{
	0xff << 0,
	0xff << 8,
	0xff << 16,
	0xff << 24,
	0xff << 32,
	0xff << 48,
	0xff << 56,
}

// calculating thes is very unreadable
// just the top 8N bits representing the unmatched bits of the hash at level N
var masks = [...]uint64{
	0xffffffffffffffff,
	0xffffffffffffff00,
	0xffffffffffff0000,
	0xffffffffff000000,
	0xffffffff00000000,
	0xffffff0000000000,
	0xffff000000000000,
	0xff00000000000000,
}

// BTW: stuff is uint8 since that saves space
// so don't make bit_length > 255 :)

// for every bit pattern of 8 bits [0-255] we have a distance [0-8] and some number of patterns
// the third dimension is not constant so we'll use slices
// maybe some kind of pointer and allocate once? append() will overallocate I guess
// we could also save space here with mirroring thing but IDC
var distance_table [size][bit_length + 1][]uint8

var hamming [size][size]uint8

// setup a lookup table for all hamming distances 0-255 x 0-255
// (could make this smaller by mirroring etc, but why bother?)
func init() {

	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {

			count := 0
			for d := i ^ j; d > 0; count++ {
				d &= d - 1 // clear the least significant bit set
			}

			hamming[i][j] = uint8(count)
			//			fmt.Printf("hamming_distance for %v,%v = %d\n", i, j, count)
			distance_table[i][count] = append(distance_table[i][count], uint8(j))
		}
	}

	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
		}
	}

	//fmt.Println(distance_table)
}

// calculate HDs using a table lookup
// todo: check if this is actually faster than the one in util
// (I think it is, but not much)
func hamming_distance(a uint64, b uint64) (distance uint8) {

	distance = 0

	var a_chunk, b_chunk uint8
	// a>0 || b>0 in this case a^b is sufficient since 0 means a=b=0 or a==b which makes their distance 0 :)
	for a^b != 0 {

		a_chunk = uint8((size - 1) & a)
		b_chunk = uint8((size - 1) & b)
		distance += hamming[a_chunk][b_chunk]
		a = a >> bit_length
		b = b >> bit_length
	}

	return
}

type SimStore struct {
	values []entry
	nodes  map[uint8]*SimStore // all subtrees based on the first bits_per_key LSB (maybe mae this an array, maybe faster?)
	level  uint8               // determines which bitrange we pick to split keys into nodes
}

type entry struct {
	key uint64
	id  int64
}

// Creates a new SimStore
func NewSimStore() *SimStore {
	return &SimStore{level: 0}
}

// Inserts a new value in the store
func (s *SimStore) Insert(text string, id int64) {
	s.insert(entry{key: SimHash(text), id: id})
}

// inserts a new value in the store, doesn't rehash etc
func (s *SimStore) insert(item entry) {

	if len(s.nodes) > 0 {

		// get the byte for this level
		b := uint8((level_chunks[s.level] & item.key) >> (s.level * bits_per_key)) // this gets you the Nth byte
		//		fmt.Printf("Node insert: level %d, getting byte: 0b%064b & 0b%064b = 0b%08b (%d)\n", s.level, level_chunks[ s.level ], item.key, b, b)
		_, exists := s.nodes[b]
		if !exists {
			s.nodes[b] = &SimStore{level: s.level + 1}
		}
		s.nodes[b].insert(item)
	} else {
		s.values = append(s.values, item)
		if len(s.values) > max_keys_per_node { // different constant here would be better I think
			s.split()
		}
	}

}

// go ever every key and put it in a node based on the value of its Nth byte
func (s *SimStore) split() {

	s.nodes = make(map[uint8]*SimStore, size)

	for _, item := range s.values {
		b := uint8((level_chunks[s.level] & item.key) >> (s.level * bits_per_key)) // this gets you the Nth byte
		_, exists := s.nodes[b]
		if !exists {
			s.nodes[b] = &SimStore{level: s.level + 1}
		}
		// don't bother with Insert(), we are splitting so we'll always be adding to the keys at this point
		s.nodes[b].values = append(s.nodes[b].values, item)
	}

	// we don't need our values anymore
	s.values = s.values[0:0]

}

// Returns a tree of nodes and number of keys per node
func (s *SimStore) String() string {
	return s.pretty("")
}

func (s *SimStore) pretty(indent string) string {
	out := ""

	if len(s.nodes) > 0 {
		out += fmt.Sprintf("%slevel % 2d\n", indent, s.level)
		for index, node := range s.nodes {
			out += fmt.Sprintf("%s%03d: %s", indent, index, node.pretty(indent+"   "))
		}
	} else {
		return fmt.Sprintf("%skeys [%d/%d]\n", indent, len(s.values), size)
	}

	return out
}

// return the number of keys and nodes in the store
func (s *SimStore) Stats() (keys, nodes int) {

	for _, subtree := range s.nodes {
		k, n := subtree.Stats()
		keys += k
		nodes += n
	}

	nodes += len(s.nodes)
	keys += len(s.values)

	return
}

// returns true if target is present in the store
func (s *SimStore) Contains(text string) (present bool, index int64) {
	return s.contains(SimHash(text))
}

// returns true if target is present in the store
func (s *SimStore) contains(target uint64) (present bool, index int64) {

	if len(s.nodes) > 0 {
		b := uint8((level_chunks[s.level] & target) >> (s.level * bits_per_key)) // this gets you the Nth byte
		subtree, exists := s.nodes[b]
		if exists {
			return subtree.contains(target)
		} else {
			return false, -1
		}
	} else {
		// check every key
		for _, item := range s.values {
			if item.key == target {
				return true, item.id
			}
		}
	}

	return false, -1
}

// returns all the hashes with a Hamming Distance of distance or less
// (less than or equal to make searching for 0 more natural)
// Checks every single key (more of a debuf thing I guess)
func (s *SimStore) FindScanAll(target uint64, distance uint8) (found []uint64) {

	found = make([]uint64, 0)

	if len(s.nodes) > 0 {
		for _, subtree := range s.nodes {
			found = append(found, subtree.FindScanAll(target, distance)...)
		}
	} else {
		for _, item := range s.values {
			if hamming_distance(item.key, target) <= distance {
				found = append(found, item.key)
			}
		}
	}

	return
}

// returns all the hashes with a Hamming Distance of distance or less
// (less than or equal to make searching for 0 more natural)
// returns the matches found as well as the number of keys and nodes checked
func (s *SimStore) Find(text string, distance uint8) (found []int64, keys_checked int, nodes_checked int) {

	return s.find(SimHash(text), distance)
}

// returns all the hashes with a Hamming Distance of distance or less
// (less than or equal to make searching for 0 more natural)
// returns the matches found as well as the number of keys and nodes checked
func (s *SimStore) find(target uint64, distance uint8) (found []int64, keys_checked int, nodes_checked int) {

	found = make([]int64, 0)
	b := int((level_chunks[s.level] & target) >> (s.level * bits_per_key)) // this gets you the Nth byte
	//	fmt.Printf("Target 0b%064b, distance %02d, byte 0b%08b [%d]\n", target, distance, b, b)

	if len(s.nodes) > 0 {

		// for all bytes that are within distance (with a max of 8) we check all nodes
		// since hamming_distance is additive
		// example: looking for 101011 with distance 2
		// we take the LSBs 11, and find everything that is within distance 2:
		// (00, 10, 01, 11) and recurse:
		// node[00].Find( 101011, 0 ) (already 'spent' distance 2)
		// node[10].Find( 101011, 1)
		// node[01].Find( 101011, 1)
		// node[11].Find( 101011, 2) (distance for this subrange was 0, 2 left to 'spend')
		//		fmt.Println(distance_table[b])

		end := min(8, distance)
		//		fmt.Printf("Searching from 0 to %03d\n", end)

		for i := uint8(0); i <= end; i++ { // check everything withing distance range
			//			fmt.Printf("Distance %d\n", i)
			for _, d := range distance_table[b][i] { // lookup which bytes are that distance from us
				subtree, exists := s.nodes[d] // check in those subtrees, if they exist
				//				fmt.Printf("Checking byte 0b%08b - node exists: %v\n", d, exists)
				if exists {
					// recurse, but the distance gets smaller
					f, k, n := subtree.find(target, distance-i)
					keys_checked += k
					nodes_checked += n
					found = append(found, f...)
				}
			}
		}

		nodes_checked += len(s.nodes)

	} else {
		// we need the part of the hash that has not been matched yet, so the (64 - bits_per_key*(level+1)) MSBs
		// eg to get the top 12 bits we do 1<<12 (0b1000000000000), -1 (0b0111111111111), then shifted to the MSBs
		// ehr, so let's just use a lookup ;)
		//mask := ()(1 << (64 - bits_per_key*(s.level+1))) -1) << (64-s.level*bits_per_key)
		mask := masks[s.level]
		masked_target := target & mask
		for _, item := range s.values {
			if hamming_distance(item.key&mask, masked_target) <= distance {
				found = append(found, item.id)
			}
		}
		keys_checked += len(s.values)
	}

	return
}

// Find the closest thing matching the input
func (s *SimStore) FindClosest(text string) int64 {

	target := SimHash(text)

	fmt.Printf("FC 0b%064b\n", target)

	// we're just going to use the target and then flip every bit in turn
	// and call contains() a lot (which is reasonably efficient)
	// in this 'tried' map we keep track of things we already tried
	// if we have very few items in the store, we'll end up wasting huge amounts
	// of time calling contains() for things that won't be found.
	// Yeah, so after trying that I quickly realized we're wasting huge amounts of space
	// storing all the keys we're trying, and also generating lots of dupe keys
	// That is a truly awful method.

	// So instead, I'll start by searching any nodes that have HD0 for the byte of our level
	// This is basically the same as contains(), but we'll try increasing byte distances
	// What happens if we get something like (8 bite example, 2 bit splits)
	// target: 00 00 00 11
	// searching depth first: 00 00 00 00, hey, distance 2, pretty close
	// however 10 00 00 11 was distance 1 but we never checked
	// but if we search breadth-first we end up searching everything :(
	// Maybe using a stack with (distance_so_far, subtree) sorted by least dist
	// and then expanding that?
	// A traditional approach to searching if you will :)
	// instead of sorting we should of course use a heap

	// ok, so let's start by checking keys if we have it (just return the one with the smallest distance?)
	// but what if other search directions find better matches? We need to also check all the frontiers
	// with HD < HD_of_first_key_found to see if we can do better.

	// first of all, do we even have nodes, bro?
	if len(s.nodes) == 0 {
		_, closest := find_closest_in_keys(&s.values, target)
		return closest
	}

	// the hard case is much much harder than the simple one unfortunately
	sh := &SearchHeap{}
	heap.Init(sh)

	// first, stick all our nodes in

	b := uint8((level_chunks[s.level] & target) >> (s.level * bits_per_key)) // this gets you the Nth byte
	for prefix, subtree := range s.nodes {

		item := &Distance{
			hamming_distance: hamming[b][prefix],
			subtree:          subtree,
		}
		heap.Push(sh, item)
	}

	// then expand the shortest distance until we hit a key
	var best_key uint64
	var closest int64
	paths_tried := 0
	// TODO: the above might be better off as an entry type

	for sh.Len() > 0 {

		shortest := heap.Pop(sh).(*Distance)
		paths_tried++
		fmt.Printf("Checking distance %03d (level: %v)\n", shortest.hamming_distance, shortest.subtree.level)
		if len(shortest.subtree.nodes) == 0 {
			fmt.Printf("This node has keys\n")
			best_key, closest = find_closest_in_keys(&shortest.subtree.values, target)
			break
		}
		// add expanded subtrees to heap
		b := uint8((level_chunks[shortest.subtree.level] & target) >> (shortest.subtree.level * bits_per_key)) // this gets you the Nth byte
		for prefix, subtree := range shortest.subtree.nodes {
			//			fmt.Printf("Expanding distance %d with %d\n", shortest.hamming_distance, hamming[b][prefix])
			item := &Distance{
				hamming_distance: shortest.hamming_distance + hamming[b][prefix], // here we add the distance, since we're going down a level
				subtree:          subtree,
			}
			heap.Push(sh, item)
		}
	}

	fmt.Printf("First key (upper bound): 0b%064b (id %v)\n", best_key, closest)

	// expand remaining nodes that have a distance < that of the first key, they might have better results
	// while the rest can't have
	upper_bound := hamming_distance(target, best_key)

	// if we found an exact match, we can skip everything else and just return that
	if upper_bound == 0 {
		return closest
	}

	fmt.Printf("Checking remaining nodes, upper bound is %d\n", upper_bound)
	for sh.Len() > 0 {
		shortest := heap.Pop(sh).(*Distance)
		paths_tried++

		// we're done if no items remain with a total HD of less than/equal the key we have
		if upper_bound <= shortest.hamming_distance {
			fmt.Printf("Every possible other path will result in larger distance, bailing out (%d paths left)\n", sh.Len())
			break
		}
		// now expand this one, until we find keys
		if len(shortest.subtree.nodes) == 0 {
			possible_closer_key, possible_closer_id := find_closest_in_keys(&shortest.subtree.values, target)
			possible_distance := hamming_distance(target, possible_closer_key)
			// woot, improvement
			if possible_distance < upper_bound {
				fmt.Printf("Got an improvement from %d to %d\n", upper_bound, possible_distance)
				upper_bound = possible_distance
				best_key = possible_closer_key
				closest = possible_closer_id
			}
		}
		// just expand nodes (if we have any)
		b := uint8((level_chunks[shortest.subtree.level] & target) >> (shortest.subtree.level * bits_per_key)) // this gets you the Nth byte
		//		fmt.Printf("Expanding distance %d \n", shortest.hamming_distance )
		for prefix, subtree := range shortest.subtree.nodes {
			new_distance := shortest.hamming_distance + hamming[b][prefix]
			// no point in adding nodes that never could lead to an improvement
			if new_distance >= upper_bound {
				continue
			}
			item := &Distance{
				hamming_distance: new_distance, // here we add the distance, since we're going down a level
				subtree:          subtree,
			}
			heap.Push(sh, item)
		}

	}

	fmt.Printf("Checked %d paths\n", paths_tried)

	// so if we have any IDs with value 0, we're returning bad things
	return closest
}

func find_closest_in_keys(v *[]entry, target uint64) (key uint64, closest int64) {
	// just check all the keys.
	distance := uint8(255) // any real one will be less
	for _, item := range *v {
		dist := hamming_distance(item.key, target)
		if dist < distance {
			distance = dist
			closest = item.id
			key = item.key
		}
	}
	return
}

// Here is our heap of HDs and SimStore nodes
// These names are awful
type Distance struct {
	hamming_distance uint8     // this is always 0-256
	subtree          *SimStore // the node we're referring to

	// below is copied from the docs, kind ugly stuff
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// this is the thing that implements the heap
// (copied from the example, no clue how the Push/Pop work)
type SearchHeap []*Distance

func (sh SearchHeap) Len() int { return len(sh) }

func (sh SearchHeap) Less(i, j int) bool {
	// We want Pop to give us the lowest hamming_distance
	return sh[i].hamming_distance < sh[j].hamming_distance
}

func (sh SearchHeap) Swap(i, j int) {
	sh[i], sh[j] = sh[j], sh[i]
	sh[i].index = i
	sh[j].index = j
}

func (sh *SearchHeap) Push(x interface{}) {
	n := len(*sh)
	item := x.(*Distance)
	item.index = n
	*sh = append(*sh, item)
}

func (sh *SearchHeap) Pop() interface{} {
	old := *sh
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*sh = old[0 : n-1]
	return item
}

// end of heap stuff

// so math.Min want float64s... and the casting is ugly
func min(a uint8, b uint8) uint8 {
	if a > b {
		return b
	}

	return a
}
