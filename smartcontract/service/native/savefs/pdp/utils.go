package pdp

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

func CalcHash(data []byte) []byte {
	var hash [sha256.Size]byte

	sha := sha256.New()
	sha.Write(data)
	sha.Sum(hash[:0])

	return hash[:]
}

func InitNodeWithHash(hash []byte, index uint64) *MerkleNode {
	return &MerkleNode{
		Layer: 0,
		Index: index,
		Hash:  hash,
	}
}

func InitNodeWithData(data []byte, index uint64) *MerkleNode {
	return InitNodeWithHash(CalcHash(data), index)
}

func copyMerkleNode(node *MerkleNode) *MerkleNode {
	newNode := &MerkleNode{
		Layer: node.Layer,
		Index: node.Index,
		Hash:  make([]byte, len(node.Hash)),
	}
	copy(newNode.Hash, node.Hash)
	return newNode
}

func MerkleHash(nodeA *MerkleNode, nodeB *MerkleNode) (*MerkleNode, error) {
	if nodeA == nil || nodeB == nil {
		return nil, fmt.Errorf("empty node")
	}
	if nodeA.Layer != nodeB.Layer {
		return nil, fmt.Errorf("node not in same layer")
	}
	if nodeA.Layer > MAX_LAYER {
		return nil, fmt.Errorf("larger than max layer")
	}
	if nodeA.Index+1 != nodeB.Index {
		return nil, fmt.Errorf("index error")
	}

	newNode := &MerkleNode{
		Layer: nodeA.Layer + 1,
		Index: nodeA.Index / 2,
		Hash:  make([]byte, sha256.Size),
	}

	nodeData := new(bytes.Buffer)
	err := nodeA.Serialize(nodeData)
	if err != nil {
		return nil, fmt.Errorf("nodeA serialize error")
	}

	err = nodeB.Serialize(nodeData)
	if err != nil {
		return nil, fmt.Errorf("nodeB serialize error")
	}

	newNode.Hash = CalcHash(nodeData.Bytes())
	return newNode, nil
}

func VerifyMerklePath(path *MerklePath, index uint64, data []byte, rootHash []byte) error {
	if path.PathLen == 0 {
		return fmt.Errorf("[VerifyMerklePath] path is 0")
	}

	// check root has match
	rootNode, err := path.GetMerkleRoot()
	if err != nil {
		return fmt.Errorf("[VerifyMerklePath] GetMerkleRoot error %s", err)
	}
	if !bytes.Equal(rootHash, rootNode.Hash) {
		return fmt.Errorf("[VerifyMerklePath] root hash mismatch, %s != %s", rootHash, rootNode.Hash)
	}

	// check data hash and index match
	node := path.Path[0]
	if node.Index != index {
		return fmt.Errorf("[VerifyMerklePath] index mismatch, %d != %d", node.Index, index)
	}

	if !bytes.Equal(CalcHash(data), node.Hash) {
		return fmt.Errorf("[VerifyMerklePath] data hash mismatch")
	}

	for i := uint64(1); i < path.PathLen-1; i++ {
		currNode := path.Path[i]
		// check the order for hash calculation
		if node.Index%2 == 0 {
			node, err = MerkleHash(node, currNode)
		} else {
			node, err = MerkleHash(currNode, node)
		}

		if err != nil {
			return fmt.Errorf("[VerifyMerklePath] MerkleHash error %s", err)
		}
	}

	if !bytes.Equal(node.Hash, rootHash) {
		return fmt.Errorf("[VerifyMerklePath] Merkel verify error")
	}

	return nil
}
