package balancer

import (
	"net"
	"net/url"
	"time"
)

type nodeState int

// nodeState constants
const (
	Healthy nodeState = iota
	Unhealthy
)

const rttHistorySize = 4

// Node is a struct representing balancing node internal state
type Node struct {
	Key     string
	URL     string
	address string // host and port without schema
	weight  uint64
	stats   rttStats
	state   nodeState
	count   uint64
}

func (n *Node) resolveAddress() {
	if _, _, err := net.SplitHostPort(n.URL); err == nil {
		n.address = n.URL
	} else {
		u, err := url.Parse(n.URL)
		if err == nil {
			n.address = u.Host
		}
	}
}

func newNode(key, url string) *Node {
	n := &Node{
		Key:   key,
		URL:   url,
		state: Healthy,
	}
	n.resolveAddress()
	//	n.stats.Add(maxResponseTime)
	return n
}

// NodesByKey is a map of Node instances by Node.Key
type NodesByKey map[string]*Node

// Nodes is a slice of Node
type Nodes []*Node

// indexOf finds element's index in slice.
func (n Nodes) indexOf(key string) int {
	for i, node := range n {
		if node.Key == key {
			return i
		}
	}
	return -1
}

// removeByIndex removes slice element by index.
func (n Nodes) removeByIndex(i int) Nodes {
	copy(n[i:], n[i+1:])
	n[len(n)-1] = nil
	return n[:len(n)-1]
}

func (n Nodes) Len() int {
	return len(n)
}

func (n Nodes) Less(i, j int) bool {
	return n[i].stats.avg < n[j].stats.avg
}

func (n Nodes) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

type rttStats struct {
	current time.Duration
	avg     time.Duration
}

func (h *rttStats) Current() time.Duration {
	return h.current
}

func (h *rttStats) Add(rtt time.Duration) {
	if h.current == 0 {
		h.avg = rtt
	} else {
		h.avg = time.Duration(weightedAverage(uint64(h.avg), EXP_5, uint64(rtt)))
	}
	h.current = rtt
}

// Weighted average implementation as in Linux load average
// http://www.perfdynamics.com/CMG/CMGslides4up.pdf

// FSHIFT
const FSHIFT = 11

// FIXED_1
const FIXED_1 = 1 << FSHIFT

// EXP_1
const EXP_1 = 1884

// EXP_5
const EXP_5 = 2014

// EXP_15
const EXP_15 = 2037

func weightedAverage(last, exp, current uint64) uint64 {
	value := last * exp
	value += current * (FIXED_1 - exp)
	value >>= FSHIFT
	return value
}
