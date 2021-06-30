package pdp

import (
	"crypto/rand"
	"testing"
)

func TestMerkleTree(t *testing.T) {
	nodeNum := 100

	data := make([][]byte, 0)
	leaves := make([]*MerkleNode, 0)
	for i := 0; i < nodeNum; i++ {
		ranData := make([]byte, 100)
		rand.Read(ranData)
		data = append(data, ranData)

		node := InitNodeWithData(ranData, uint64(i))
		leaves = append(leaves, node)
	}

	tree, err := InitMerkleTree(leaves)
	if err != nil {
		t.Fatalf("initTree error %s\n", err)
		return
	}
	//tree.PrintTree()

	root, err := tree.GetMerkleRoot()
	if err != nil {
		t.Fatalf("GetMerkleRoot error %s\n", err)
		return
	}

	//t.Logf("root : %+v\n", root)

	for index := uint64(0); index < uint64(nodeNum); index++ {
		path, err := tree.GetMerklePath(index)
		if err != nil {
			t.Fatalf("GetMerklePath error %s\n", err)
			return
		}

		//t.Logf("index %d\n", index)
		//path.PrintPath()

		err = VerifyMerklePath(path, index, data[index], root.Hash)
		if err != nil {
			t.Fatalf("VerifyMerklePath error %s\n", err)
			return
		}
		//t.Logf("Verify Ok")
	}

}
