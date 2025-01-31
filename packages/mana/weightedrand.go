package mana

import (
	"math/rand"
	"sort"
)

// RandChoice is a generic wrapper that can be used to add weights for any item.
type RandChoice struct {
	Item   interface{}
	Weight int
}

// NewChoice creates a new RandChoice with specified item and weight.
func NewChoice(item interface{}, weight int) RandChoice {
	return RandChoice{Item: item, Weight: weight}
}

// A RandChooser caches many possible Choices in a structure designed to improve
// performance on repeated calls for weighted random selection.
type RandChooser struct {
	data   []RandChoice
	totals []int
	max    int
}

// NewRandChooser initializes a new RandChooser for picking from the provided Choices.
func NewRandChooser(cs ...RandChoice) *RandChooser {
	sort.Slice(cs, func(i, j int) bool {
		return cs[i].Weight < cs[j].Weight
	})
	totals := make([]int, len(cs))
	runningTotal := 0
	for i, c := range cs {
		runningTotal += c.Weight
		totals[i] = runningTotal
	}
	return &RandChooser{data: cs, totals: totals, max: runningTotal}
}

// Pick returns N weighted random items from the RandChooser.
//
// Utilizes global rand as the source of randomness -- you will likely want to seed it.
func (chs *RandChooser) Pick(n uint) []interface{} {
	rands := rand.Perm(chs.max)
	var res []interface{}
	for _, r := range rands {
		r++
		i := sort.SearchInts(chs.totals, r)
		if i > len(chs.data)-1 {
			i = len(chs.data) - 1
		}
		res = append(res, chs.data[i].Item)
		if len(res) == int(n) {
			break
		}
		chs.remove(i)
		if int(n) <= len(res) {
			return res[:n]
		}
	}
	return res
}

// remove picked element at index
func (chs *RandChooser) remove(i int) {
	tmp := chs.data
	tmp[i] = tmp[len(tmp)-1]
	tmp[len(tmp)-1] = RandChoice{}
	tmp = tmp[:len(tmp)-1]

	tmpChooser := NewRandChooser(tmp...)
	*chs = *tmpChooser
}
