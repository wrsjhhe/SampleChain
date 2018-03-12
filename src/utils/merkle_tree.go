package utils

import "crypto/sha256"

type MerkleTree struct {
	RootNode   *MerkleNode
}

type MerkleNode struct {
	Left       *MerkleNode
	Right      *MerkleNode
	Data       []byte
}

func NewMerkleTree(data [][]byte)*MerkleTree  {
	var nodes  []MerkleNode

	if len(data)%2 !=0 {
		data = append(data,data[len(data)-1])
	}

	for _,datum:=range data{
		var node = NewMerkleNode(nil,nil,datum)
		nodes = append(nodes,*node)
	}

	for i:=0;i<len(data)/2;i++{
		var newLevel   []MerkleNode

		for j:=0;j<len(nodes);j+=2{
			var node = NewMerkleNode(&nodes[j],&nodes[j+1],nil)
			newLevel = append(newLevel,*node)
		}
		nodes = newLevel
	}

	var mTree = MerkleTree{&nodes[0]}

	return &mTree
}

func NewMerkleNode(left,right *MerkleNode,data []byte)*MerkleNode  {
	var mNode = MerkleNode{}

	if left == nil && right == nil{
		var hash = sha256.Sum256(data)
		mNode.Data = hash[:]
	}else{
		var preHashes = append(left.Data,right.Data...)
		var hash = sha256.Sum256(preHashes)
		mNode.Data = hash[:]
	}

	mNode.Left = left
	mNode.Right = right

	return &mNode
}
