package balancer

import "time"

// NodeStat represents node statistics info
type NodeStat struct {
	Key            string        `json:"key"`
	Value          string        `json:"value"`
	RolloutType    string        `json:"rollout_type"`
	Healthy        bool          `json:"healthy"`
	HitProbability float64       `json:"hitProbability"`
	HitCount       uint64        `json:"hitCount"`
	RTT            time.Duration `json:"rtt"`
	RTTAverage     time.Duration `json:"rttAverage"`
	Connected      bool          `json:"connected"` // Connected is true if connection is established (for GRPC)
}

// StatsByKey is a sorting type
type StatsByKey []NodeStat

func (s StatsByKey) Len() int {
	return len(s)
}

func (s StatsByKey) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s StatsByKey) Less(i, j int) bool {
	return s[i].Key < s[j].Key
}

// StatsByValue is a sorting type
type StatsByValue []NodeStat

func (s StatsByValue) Len() int {
	return len(s)
}

func (s StatsByValue) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s StatsByValue) Less(i, j int) bool {
	return s[i].Value < s[j].Value
}

// StatsByProbability is a sorting type
type StatsByProbability []NodeStat

func (s StatsByProbability) Len() int {
	return len(s)
}

func (s StatsByProbability) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s StatsByProbability) Less(i, j int) bool {
	return s[i].HitProbability < s[j].HitProbability
}
