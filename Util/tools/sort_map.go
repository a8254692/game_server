package tools

import "sort"

func RankByCount(wordFrequencies map[uint32]uint32) PairList {
	pl := make(PairList, len(wordFrequencies))
	i := 0
	for k, v := range wordFrequencies {
		pl[i] = Pair{k, v}
		i++
	}
	//从小到大排序
	//sort.Sort(pl)
	//从大到小排序
	sort.Sort(sort.Reverse(pl))
	return pl
}

type Pair struct {
	Key   uint32
	Value uint32
}

type PairList []Pair

func (p PairList) Len() int {
	return len(p)
}

func (p PairList) Less(i, j int) bool {
	return p[i].Value < p[j].Value
}

func (p PairList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

type MapSorter []SortItem

type SortItem struct {
	Key string      `json:"key"`
	Val interface{} `json:"val"`
}

func (ms MapSorter) Len() int {
	return len(ms)
}
func (ms MapSorter) Less(i, j int) bool {
	return ms[i].Key < ms[j].Key // 按键排序
	// return ms[i].Value <ms[j].Value //按值排序
}
func (ms MapSorter) Swap(i, j int) {
	ms[i], ms[j] = ms[j], ms[i]
}

func MapSort(m map[string]string) []map[string]string {
	ms := make(MapSorter, 0, len(m))

	for k, v := range m {
		ms = append(ms, SortItem{k, v})
	}
	sort.Sort(ms)
	result := make([]map[string]string, 0)
	for _, p := range ms {
		result = append(result, map[string]string{p.Key: p.Val.(string)})
	}
	return result
}
