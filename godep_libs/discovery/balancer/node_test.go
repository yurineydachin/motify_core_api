package balancer

import (
	"reflect"
	"testing"
)

func TestNodes_RemoveByIndex(t *testing.T) {
	node1 := &Node{Key: "1", URL: "1"}
	node2 := &Node{Key: "2", URL: "2"}
	node3 := &Node{Key: "3", URL: "3"}
	nodes := Nodes{node1, node2, node3}

	nodes = nodes.removeByIndex(1)
	assertNodesEqual(t, nodes, Nodes{node1, node3})

	nodes = nodes.removeByIndex(0)
	assertNodesEqual(t, nodes, Nodes{node3})

	nodes = nodes.removeByIndex(0)
	assertNodesEqual(t, nodes, Nodes{})
}

func TestNodes_IndexOf(t *testing.T) {
	node1 := &Node{Key: "1", URL: "1"}
	node2 := &Node{Key: "2", URL: "2"}
	node3 := &Node{Key: "3", URL: "3"}
	nodes := Nodes{node1, node2, node3}

	i := nodes.indexOf("1")
	if i != 0 {
		t.FailNow()
	}

	i = nodes.indexOf("2")
	if i != 1 {
		t.FailNow()
	}

	i = nodes.indexOf("3")
	if i != 2 {
		t.FailNow()
	}

	i = nodes.indexOf("42")
	if i != -1 {
		t.FailNow()
	}
}

func TestResolveAddress(t *testing.T) {
	tests := []struct {
		In  string
		Out string
	}{
		{In: "", Out: ""},
		{In: "test", Out: ""},
		{In: "127.0.0.1", Out: ""},
		{In: "localhost", Out: ""},
		{In: "localhost:1234", Out: "localhost:1234"},
		{In: "127.0.0.1:8080", Out: "127.0.0.1:8080"},
		{In: "http://localhost:1234", Out: "localhost:1234"},
		{In: "http://127.0.0.1:8080", Out: "127.0.0.1:8080"},
		{In: "ftp://127.0.0.1:8080", Out: "127.0.0.1:8080"},
	}
	for _, c := range tests {
		n := &Node{Key: "1", URL: c.In}
		n.resolveAddress()
		if n.address != c.Out {
			t.Errorf("expected address for URL: '%s' is '%s', got: '%s'", n.URL, c.Out, n.address)
		}
	}
}

func assertNodesEqual(t *testing.T, actual Nodes, expected Nodes) {
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("actual array (%v) not equal to expected array (%v)", actual, expected)
	}
}
