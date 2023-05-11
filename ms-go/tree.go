package ms_go

import "strings"

type treeNode struct {
	name       string
	children   []*treeNode
	routerName string
}

func (t *treeNode) Put(path string) {
	curNode := t
	strs := strings.Split(path, "/")
	for index, name := range strs {
		if index == 0 {
			continue
		}
		needAdd := true
		for _, node := range curNode.children {
			if node.name == name {
				needAdd = false
				curNode = node
				break
			}
		}
		if needAdd {
			child := &treeNode{
				name:     name,
				children: make([]*treeNode, 0),
			}
			curNode.children = append(curNode.children, child)
			curNode = child
		}
	}
}

func (t *treeNode) Get(path string) *treeNode {
	curNode := t
	strs := strings.Split(path, "/")
	routerName := ""
	for index, name := range strs {
		if index == 0 {
			continue
		}
		isContinue := false
		for _, node := range curNode.children {
			if node.name == name || node.name == "*" || strings.Contains(node.name, ":") {
				routerName += "/" + node.name
				if index == len(strs)-1 {
					node.routerName = routerName
					return node
				} else {
					curNode = node
					isContinue = true
					break
				}
			}
		}
		if !isContinue {
			return nil
		}
	}
	return nil
}
