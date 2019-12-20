package raft

import "sort"

type sortPair struct {
	key int
	value int
}

type sortPairList []sortPair

func (spl sortPairList) Len() int {
	return len(spl)
}

func (spl sortPairList) Less(i, j int) bool {
	return spl[i].value < spl[j].value
}

func (spl sortPairList) Swap(i, j int) {
	spl[i], spl[j] = spl[j], spl[i]
}

func sortMapByValue(m map[int]int) sortPairList {
	spl := make(sortPairList, len(m))
	i := 0
	for k, v := range m {
		spl[i] = sortPair{
			key:   k,
			value: v,
		}
		i++
	}
	sort.Sort(spl)
	return spl
}
