package pdp

import (
	"fmt"
	"github.com/saveio/themis/smartcontract/service/native/utils"
	"io"
)

const (
	MAX_LAYER = 1000
)

type MerkleNode struct {
	Layer uint64
	Index uint64
	Hash  []byte // store hash even for leaves nodes to reduce memory use for large file
}

func (this *MerkleNode) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.Layer); err != nil {
		return fmt.Errorf("[MerkleNode] Layer serialize error:%v", err)
	}
	if err := utils.WriteVarUint(w, this.Index); err != nil {
		return fmt.Errorf("[MerkleNode] Index serialize error:%v", err)
	}
	if err := utils.WriteBytes(w, this.Hash); err != nil {
		return fmt.Errorf("[MerkleNode] Hash serialize error:%v", err)
	}
	return nil
}

func (this *MerkleNode) Deserialize(r io.Reader) error {
	var err error
	if this.Layer, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[MerkleNode] Layer deserialize error:%v", err)
	}
	if this.Index, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[MerkleNode] Index deserialize error:%v", err)
	}
	if this.Hash, err = utils.ReadBytes(r); err != nil {
		return fmt.Errorf("[MerkleNode] Index deserialize error:%v", err)
	}
	return nil
}

type MerklePath struct {
	PathLen uint64
	Path    []*MerkleNode
}

func (this *MerklePath) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.PathLen); err != nil {
		return fmt.Errorf("[MerklePath] PathLen deserialize error:%v", err)
	}
	for i := uint64(0); i < this.PathLen; i++ {
		if err := this.Path[i].Serialize(w); err != nil {
			return fmt.Errorf("[MerklePath] Path deserialize error:%v", err)
		}
	}
	return nil
}

func (this *MerklePath) Deserialize(r io.Reader) error {
	var err error
	if this.PathLen, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[MerklePath] PathLen deserialize error:%v", err)
	}

	nodes := make([]*MerkleNode, 0)
	for i := uint64(0); i < this.PathLen; i++ {
		node := new(MerkleNode)
		if err = node.Deserialize(r); err != nil {
			return fmt.Errorf("[MerklePath] Pathdeserialize error:%v", err)
		}
		nodes = append(nodes, node)
	}
	this.Path = nodes
	return nil
}

func (this *MerklePath) GetMerkleRoot() (*MerkleNode, error) {
	if this.PathLen == 0 {
		return nil, fmt.Errorf("[GetMerkleRoot] path len is 0")
	}
	return this.Path[this.PathLen-1], nil
}

func (this *MerklePath) PrintPath() {
	fmt.Printf("path len : %d\n", this.PathLen)
	for i := uint64(0); i < this.PathLen; i++ {
		fmt.Printf("%+v\n", this.Path[i])
	}
	fmt.Println()
}

type MerkleTree struct {
	Layers [][]*MerkleNode
}

func InitMerkleTree(leaves []*MerkleNode) (*MerkleTree, error) {
	if len(leaves) == 0 {
		return nil, fmt.Errorf("[InitMerkleTree] leaves empty")
	}

	tree := &MerkleTree{Layers: make([][]*MerkleNode, 0)}
	tree.Layers = append(tree.Layers, make([]*MerkleNode, 0))
	// check index is in sequence and init layer 0
	for i, node := range leaves {
		if node.Layer != 0 || uint64(i) != node.Index {
			return nil, fmt.Errorf("[InitMerkleTree] layer or index error for leave node")
		}
		tree.Layers[0] = append(tree.Layers[0], copyMerkleNode(node))
	}

	depth := 0
	for {
		layerHashes := tree.Layers[depth]
		layerHashesLen := len(layerHashes)
		if layerHashesLen == 1 {
			return tree, nil
		}

		newLayerHashes := make([]*MerkleNode, 0)
		n := layerHashesLen / 2

		for i := 0; i < n; i++ {
			newNode, err := MerkleHash(layerHashes[2*i], layerHashes[2*i+1])
			if err != nil {
				return nil, fmt.Errorf("[InitMerkleTree] compute merkleHash error")
			}
			newLayerHashes = append(newLayerHashes, newNode)
		}

		if layerHashesLen == 2*n+1 {
			node := layerHashes[2*n]
			nodeCopy := copyMerkleNode(node)
			nodeCopy.Index = uint64(2*n + 1)

			// copy node should also be inserted into the layer for merkle path calculation
			// NOTE: after append, the underlying byte slice may have been extended, so should re-assign the value
			layerHashes = append(layerHashes, nodeCopy)
			tree.Layers[depth] = layerHashes

			newNode, err := MerkleHash(node, nodeCopy)
			if err != nil {
				return nil, fmt.Errorf("[InitMerkleTree] compute merkleHash error")
			}
			newLayerHashes = append(newLayerHashes, newNode)
		}

		tree.Layers = append(tree.Layers, newLayerHashes)

		depth++
	}
	return tree, nil
}

func (this *MerkleTree) PrintTree() {
	for i, layer := range this.Layers {
		fmt.Printf("Layer %d:\n", i)
		for i, node := range layer {
			fmt.Printf("index %d:%+v\n", i, node)
		}
		fmt.Println()
	}
}

func (this *MerkleTree) isEmpty() bool {
	return len(this.Layers) == 0
}

func (this *MerkleTree) GetNodeWithIndex(index uint64) (*MerkleNode, error) {
	if this.isEmpty() {
		return nil, fmt.Errorf("GetNodeWithIndex merkle tree is empty")
	}
	if index >= uint64(len(this.Layers[0])) {
		return nil, fmt.Errorf("GetNodeWithIndex index out of bound")
	}

	return copyMerkleNode(this.Layers[0][index]), nil
}

// get merkle path with node index
func (this *MerkleTree) GetMerklePath(index uint64) (*MerklePath, error) {
	if this.isEmpty() {
		return nil, fmt.Errorf("[GetMerklePath] empty merkle tree")
	}

	leafNodesLen := uint64(len(this.Layers[0]))
	if leafNodesLen <= index {
		return nil, fmt.Errorf("[GetMerklePath] index out of bound")
	}

	layerNum := len(this.Layers)

	path := &MerklePath{
		PathLen: uint64(layerNum + 1), // node with index plus one node in each layer
		Path:    make([]*MerkleNode, 0),
	}

	// append the node in the index as the first in the path
	path.Path = append(path.Path, this.Layers[0][index])

	var neighbourIndex uint64
	layerIndex := index

	for i := 0; i < layerNum; i++ {
		layer := this.Layers[i]

		if len(layer) == 1 {
			path.Path = append(path.Path, layer[0])
			return path, nil
		}

		if layerIndex%2 == 0 {
			neighbourIndex = layerIndex + 1
		} else {
			neighbourIndex = layerIndex - 1
		}

		path.Path = append(path.Path, layer[neighbourIndex])
		layerIndex = layerIndex / 2
	}

	return path, nil
}

func (this *MerkleTree) GetMerkleRoot() (*MerkleNode, error) {
	if this.isEmpty() {
		return nil, fmt.Errorf("[GetMerkleRoot] empty merkle tree")
	}
	return this.Layers[len(this.Layers)-1][0], nil
}
