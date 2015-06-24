// Frontier
package priorityQueue

import (
	"container/heap"
)

type PriorityQueue []*URL

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i,j int) bool {
	if(pq[i].seed)  {
	return true
	}
	if(pq[j].seed) {
	return false
	}
	if(pq[i].in_links == pq[j].in_links) {
		return pq[i].index < pq[j].index 
	}
	return pq[i].in_links > pq[j].in_links
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}){
	n := len(*pq);
	item := x.(*URL)
	item.index = n
	*pq = append(*pq,item)
}

func NewPQueue() *PriorityQueue {
    return &PriorityQueue{}
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	length := len(*pq)
	item := old[length-1]
	item.index = -1
	*pq = old[0 : length-1]
	return item
}

func (pq *PriorityQueue) update(url *URL,inlinks int){
	url.in_links = url.in_links + 1
	heap.Fix(pq,url.index)
} 