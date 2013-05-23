package sim

import "testing"
import "fmt"
import "math/rand"

func TestFindClosestKeys(t *testing.T) {

	simstore := NewSimStore()
	
	simstore.Insert( "It was the best of times, it was the worst of times,", 1 )
	simstore.Insert( "it was the age of wisdom, it was the age of foolishness,", 2 )
	simstore.Insert( "it was the epoch of belief, it was the epoch of incredulity,", 3 )
	simstore.Insert( "it was the season of Light, it was the season of Darkness,", 4 )
	simstore.Insert( "it was the spring of hope, it was the winter of despair,", 5 )
	simstore.Insert( "we had everything before us, we had nothing before us,", 6 )
	simstore.Insert( "we were all going direct to Heaven, we were all going direct the other way", 7 )

	found := simstore.FindClosest("It was the best of times, it was the worst of times,")
	if found != 1 {
		t.Error("FindClosest didn't find the perfect match")
	}

	found = simstore.FindClosest("It was the best of times and it was the worst of times")
	if found != 1 {
		t.Error("FindClosest didn't find the good match")
	}

}

// this is where we insert >> 256 items so we are sure to test the FindClosest codepath that includes searching the nodes
func TestFindClosestNodes(t *testing.T) {

	simstore := NewSimStore()

	// insert 100K nonsense
	r := rand.New(rand.NewSource(1234))
	for i:=0; i<100000; i++ {
		simstore.Insert( fmt.Sprintf("%016x", r.Int63()), int64(i) )
	}
	
	// insert the thing we hope to find
	simstore.Insert( "It was the best of times, it was the worst of times,", -1 )
	
	// exact match
	found := simstore.FindClosest("It was the best of times, it was the worst of times,")
	if found != -1 {
		t.Error("FindClosest didn't find the perfect match")
	}
	
	// close match
	found = simstore.FindClosest("It was the best of times, it was peanut butter jelly time")
	if found != -1 {
		t.Error("FindClosest didn't find the perfect match")
	}
}