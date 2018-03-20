package dconfig

import (
	"sort"
	"strconv"
	"strings"

	etcdcl "github.com/coreos/etcd/client"
)

type Setting struct {
	Key         string     `json:"key"`
	Description string     `json:"description"`
	Value       string     `json:"value"`
	History     []Snapshot `json:"values"`
}

type Snapshot struct {
	key       string
	Timestamp int64  `json:"timestamp"`
	Owner     string `json:"owner"`
	Value     string `json:"value"`
}

type ByTimestamp []Snapshot

func (s ByTimestamp) Len() int {
	return len(s)
}

func (s ByTimestamp) Less(i, j int) bool {
	return s[i].Timestamp > s[j].Timestamp
}

func (s ByTimestamp) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func nodeToSetting(node *etcdcl.Node) *Setting {
	i := strings.LastIndexByte(node.Key, '/')
	name := node.Key[i+1:]
	setting := &Setting{
		Key: name,
	}

	for _, settingKeyNode := range node.Nodes {
		key := strings.TrimPrefix(settingKeyNode.Key, node.Key+"/")
		if key == "description" {
			setting.Description = settingKeyNode.Value
			continue
		}
		if key != "values" {
			continue
		}

		setting.History = make([]Snapshot, settingKeyNode.Nodes.Len())
		for i, valueNode := range settingKeyNode.Nodes {
			ts, _ := strconv.ParseInt(strings.TrimPrefix(valueNode.Key, settingKeyNode.Key+"/"), 0, 64)
			snapshot := &setting.History[i]
			snapshot.key = valueNode.Key
			// TODO: fix it. js hack
			snapshot.Timestamp = ts / 1000000
			for _, valueKeyNode := range valueNode.Nodes {
				switch strings.TrimPrefix(valueKeyNode.Key, valueNode.Key+"/") {
				case "value":
					snapshot.Value = valueKeyNode.Value
				case "owner":
					snapshot.Owner = valueKeyNode.Value
				}
			}
		}
	}

	if len(setting.History) > 0 {
		sort.Sort(ByTimestamp(setting.History))
		setting.Value = setting.History[0].Value
	}

	return setting
}
