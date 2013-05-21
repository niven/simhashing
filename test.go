package main

import "cveenboer/sim"
import "fmt"
import "math/rand"

func main() {

	store := sim.NewSimStore()
	
	r := rand.New(rand.NewSource(99))
	const N = 100
	const MAX = 1 << 32
	
	var distances = [...]uint8{1,3,5,8,16}
	
	const prefill = 20*1000//*1000
	
	// prefill some entries
	for i:=0; i<prefill; i++ {
		num := r.Int63n(MAX)
		str := fmt.Sprintf("%010d", num)
//		fmt.Printf("inserting '%s' = 0b%064b\n", str, h)
		store.Insert( str, num )
	}
	fmt.Println("Prefill done")
	for i:=prefill; i<prefill+N; i++ {
		num := r.Int63n(MAX)
		str := fmt.Sprintf("%010d", num)
//		fmt.Printf("inserting '%s' = 0b%064b\n", str, h)
		store.Insert( str, num )
		if !store.Contains( str ) {
			panic("Containment failure")
		}
		
		for _,s := range distances {
	//		found1 := store.FindScanAll( str, s )
			found2, k, n := store.Find( str, s )
//			if len(found1) != len(found2) {
//				panic("Did not find the same number of target")
//			} else {
//				fmt.Printf("Foudn scan/find (%d): %v / %v\n", s, found1, found2)
//			}
//			fmt.Println()

    		fmt.Printf("n=%05d, distance: %d, found/searched %d/%d keys = %.2f%% space (%v nodes)\n", i, s, len(found2), k, 100*float64(k)/float64(i), n)

		}

	}
//	fmt.Println( store.String() )
	k,n := store.Stats()
	fmt.Println( "keys/nodes", k, n )
}
