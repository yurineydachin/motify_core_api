package service

import (
	"godep.lzd.co/discovery/balancer"
	"godep.lzd.co/mobapi_lib/resources"
	"sort"
	"strconv"
)

type balancerResource struct {
	name string
	b    balancer.ILoadBalancer
}

func (r balancerResource) Caption() string {
	return r.name
}

func (r balancerResource) Status() resources.Status {
	stats := resources.Status{
		Header: []string{"Key", "Probability", "Hit count", "RTT last", "RTT average"},
	}
	balancerStats := r.b.Stats()
	sort.Sort(balancer.StatsByProbability(balancerStats))
	for _, bs := range r.b.Stats() {
		var l string
		if bs.Healthy {
			l = resources.ResourceStatusOK
		} else {
			l = resources.ResourceStatusFail
		}

		stats.Rows = append(stats.Rows, resources.StatusRow{
			Level: l,
			Data: []string{
				bs.Value,
				strconv.FormatFloat(bs.HitProbability*100., 'f', 2, 64) + "%",
				strconv.FormatUint(bs.HitCount, 10),
				bs.RTT.String(),
				bs.RTTAverage.String(),
			},
		})
	}
	return stats
}

type grpcResource struct {
	name     string
	balancer interface {
		Stats() []balancer.NodeStat
	}
}

func (r grpcResource) Caption() string {
	return r.name
}

func (r grpcResource) Status() resources.Status {
	stats := resources.Status{
		Header: []string{"Key", "Connected", "Probability", "Hit count", "RTT last", "RTT average"},
	}
	balancerStats := r.balancer.Stats()
	sort.Sort(balancer.StatsByProbability(balancerStats))
	for _, bs := range r.balancer.Stats() {
		l := resources.ResourceStatusFail
		if bs.Healthy && bs.Connected {
			l = resources.ResourceStatusOK
		}
		connected := resources.ResourceNotConnected
		if bs.Connected {
			connected = resources.ResourceConnected
		}
		stats.Rows = append(stats.Rows, resources.StatusRow{
			Level: l,
			Data: []string{
				bs.Value,
				connected,
				strconv.FormatFloat(bs.HitProbability*100., 'f', 2, 64) + "%",
				strconv.FormatUint(bs.HitCount, 10),
				bs.RTT.String(),
				bs.RTTAverage.String(),
			},
		})
	}
	return stats
}
