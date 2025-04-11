package helpers

import (
	"bytes"
	"sort"
)

type prefixTreeNode struct {
	fullPath []byte
	segment  []byte
	children []*prefixTreeNode // sorted by first byte
	isLeaf   bool
}

func newPrefixLeafNode(s []byte, prefixLen int) *prefixTreeNode {
	return &prefixTreeNode{
		fullPath: s,
		segment:  s[prefixLen:],
		children: make([]*prefixTreeNode, 0),
		isLeaf:   true,
	}
}

type prefixTreeSorter []*prefixTreeNode

func (a prefixTreeSorter) Len() int           { return len(a) }
func (a prefixTreeSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a prefixTreeSorter) Less(i, j int) bool { return a[i].segment[0] < a[j].segment[0] }

func (n *prefixTreeNode) add(s []byte, prefixLen int) *prefixTreeNode {
	if !bytes.HasPrefix(s[prefixLen:], n.segment) {
		// need to fork
		commonLen := 0
		for idx, b := range s[prefixLen:] {
			if b == n.segment[idx] {
				commonLen++
			} else {
				break
			}
		}
		if commonLen == len(s[prefixLen:]) {
			newN := newPrefixLeafNode(s, prefixLen)
			newN.children = append(newN.children, n)
			n.segment = n.segment[commonLen:]
			return newN
		}

		newN := &prefixTreeNode{
			segment:  n.segment[:commonLen],
			children: []*prefixTreeNode{n, newPrefixLeafNode(s, prefixLen+commonLen)},
			isLeaf:   false,
		}
		n.segment = n.segment[commonLen:]
		sort.Sort(prefixTreeSorter(newN.children))
		return newN
	}

	if len(s[prefixLen:]) == len(n.segment) {
		n.isLeaf = true
		return n
	}

	prefixLen += len(n.segment)
	idxByte := s[prefixLen]
	curPos := len(n.children) - 1
	for ; curPos >= 0; curPos-- {
		if n.children[curPos].segment[0] <= idxByte {
			break
		}
	}

	// no some prefix
	if curPos < 0 || n.children[curPos].segment[0] < idxByte {
		n.children = append(
			n.children,
			newPrefixLeafNode(s, prefixLen))
		sort.Sort(prefixTreeSorter(n.children))
		return n
	}

	n.children[curPos] = n.children[curPos].add(s, prefixLen)
	return n
}

func (n *prefixTreeNode) find(s []byte, prefixLen int) *prefixTreeNode {
	if !bytes.HasPrefix(s[prefixLen:], n.segment) {
		return nil
	}

	if len(s[prefixLen:]) == len(n.segment) {
		if n.isLeaf {
			return n
		}

		return nil
	}

	prefixLen += len(n.segment)
	idxByte := s[prefixLen]
	// bin search
	l, r := 0, len(n.children)-1
	for l <= r {
		mid := (l + r) / 2
		node := n.children[mid]
		if node.segment[0] == idxByte {
			return node.find(s, prefixLen)
		} else if node.segment[0] < idxByte {
			l = mid + 1
		} else {
			r = mid - 1
		}
	}

	return nil
}

// PrefixTree 一个用于快速查找字符串的前缀树.
//
// PrefixTree is a prefix-tree based structure for fast finding.
type PrefixTree struct {
	root *prefixTreeNode
}

// Add 添加一个字符串.
//
// Add a string into the tree.
func (t *PrefixTree) Add(s string) {
	if t.root == nil {
		t.root = &prefixTreeNode{
			children: []*prefixTreeNode{},
			isLeaf:   false,
		}
	}

	t.root.add([]byte(s), 0)
}

// Find 检查字符串是否存在.
//
// Find a string from the tree.
func (t *PrefixTree) Find(s string) bool {
	if t.root == nil {
		return false
	}

	if n := t.root.find([]byte(s), 0); n != nil {
		return true
	}

	return false
}
