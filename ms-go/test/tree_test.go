package test

import (
	"fmt"
	"github.com/gaolaoge/ms-go"
	"testing"
)

func TestTreeNode(t *testing.T) {
	root := &ms_go.treeNode{name: "/", children: make([]*ms_go.treeNode, 0)}

	root.Put("/user/create/:name")
	root.Put("/user/detail/:id")
	root.Put("/user/delete/:id")
	root.Put("/user/update/:id")
	root.Put("/user/list")

	node := root.Get("/user/create/gaoge")
	fmt.Println(node)
	node = root.Get("/user/somethings/create/gaoge")
	fmt.Println(node)
	node = root.Get("/user/detail/1")
	fmt.Println(node)
	node = root.Get("/user/delete/1")
	fmt.Println(node)
	node = root.Get("/user/update/1")
	fmt.Println(node)
	node = root.Get("/user/list")
	fmt.Println(node)

}
