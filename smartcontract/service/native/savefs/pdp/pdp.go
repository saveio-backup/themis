package pdp

import (
	"bytes"
	"fmt"
	bp "github.com/saveio/themis/crypto/bulletproof_pdp_25519"
	"sync"
)

type Pdp struct {
	lock    sync.RWMutex
	version uint64
	algo    *bp.InnerProductPDP
	trees   map[FileID]*MerkleTree
}

func NewPdp(version uint64) *Pdp {
	return &Pdp{
		version: 1,
		algo:    bp.NewInnerProductPDP(),
		trees:   make(map[FileID]*MerkleTree),
	}
}

func (this *Pdp) IsMerkleTreeExistForFile(fileId FileID) bool {
	this.lock.RLock()
	defer this.lock.RUnlock()

	_, exist := this.trees[fileId]
	return exist
}

// init merkle tree with merkles nodes, cannot use data directly since it may consume too much memory
func (this *Pdp) InitMerkleTreeForFile(fileId FileID, nodes []*MerkleNode) error {
	if this.IsMerkleTreeExistForFile(fileId) {
		return fmt.Errorf("InitMerkleTreeForFile fileId %v already init", fileId)
	}

	tree, err := InitMerkleTree(nodes)
	if err != nil {
		return fmt.Errorf("InitMerkleTree init tree error %s", err)
	}
	this.lock.Lock()
	defer this.lock.Unlock()

	this.trees[fileId] = tree
	return nil
}

func (this *Pdp) GetRootHashForFile(fileId FileID) ([]byte, error) {
	if !this.IsMerkleTreeExistForFile(fileId) {
		return nil, fmt.Errorf("GenUniqueIdWithNodes merkle tree not init")
	}
	root, err := this.trees[fileId].GetMerkleRoot()
	if err != nil {
		return nil, fmt.Errorf("GenUniqueIdWithNodes get merkle tree root error")
	}
	return root.Hash, nil
}

func (this *Pdp) GenerateTag(blocks []Block, fileId FileID) ([]Tag, error) {
	bpBlocks, err := convertBlocks(blocks)
	if err != nil {
		return nil, fmt.Errorf("GenerateTag convertBlocks error %s", err)
	}

	tags := make([]Tag, 0)
	bpTags := this.algo.GenTag(bpBlocks, fileId)
	for _, tag := range bpTags {
		t := Tag{}
		copy(t[:], tag[:])
		tags = append(tags, t)
	}
	return tags, nil
}

func (this *Pdp) GenerateProof(version uint64, blocks []Block, fileIds []FileID, challenges []Challenge) ([]byte, error) {
	bpBlocks, err := convertBlocks(blocks)
	if err != nil {
		return nil, fmt.Errorf("GenerateTag convertBlocks error %s", err)
	}
	bpChallenges := convertChallenges(challenges)
	bpFileIDs := convertFileIDs(fileIds)

	proof, err := this.algo.ProofGenerate(version, bpBlocks, bpFileIDs, bpChallenges)
	if err != nil {
		return nil, fmt.Errorf("GenerateProof error %s", err)
	}
	return proof, nil
}

func (this *Pdp) VerifyProof(version uint64, proofs []byte, fileIds []FileID, tags []Tag, challenges []Challenge) bool {
	bpTags := convertTags(tags)
	bpChallenges := convertChallenges(challenges)
	bpFileIDs := convertFileIDs(fileIds)
	return this.algo.ProofVerify(version, proofs, bpFileIDs, bpTags, bpChallenges)
}

// generate the proof with blocks with the index in the challenges, then generate the for merkle path for tag
func (this *Pdp) GenerateProofWithMerklePath(version uint64, blocks []Block, fileIds []FileID, tags []Tag, challenges []Challenge) ([]byte, []*MerklePath, error) {
	for _, fileId := range fileIds {
		if !this.IsMerkleTreeExistForFile(fileId) {
			return nil, nil, fmt.Errorf("GenerateProofWithMerklePath merkle tree not init for fileId %v", fileId)
		}
	}

	if len(blocks) == 0 {
		return nil, nil, fmt.Errorf("GenerateProofWithMerklePath length of blocks is 0")
	}
	if len(blocks) != len(tags) || len(blocks) != len(challenges) {
		return nil, nil, fmt.Errorf("GenerateProofWithMerklePath length of blocks, tags, challenges not the same")
	}

	proof, err := this.GenerateProof(version, blocks, fileIds, challenges)
	if err != nil {
		return nil, nil, fmt.Errorf("GenerateProofWithMerklePath GenerateProof error %s", err)
	}

	pathResult := make([]*MerklePath, 0)
	for i, challenge := range challenges {
		fileId := fileIds[i]

		tree := this.trees[fileId]
		// NOTE: challenge indexes is not global indexes but the indexes within the file
		// so it must be converted from global indexes with files order and file size info
		node, err := tree.GetNodeWithIndex(uint64(challenge.Index))
		if err != nil {
			return nil, nil, fmt.Errorf("GenerateProofWithMerklePath GetNodeWithIndex error %s", err)
		}

		if !bytes.Equal(node.Hash, CalcHash(tags[i][:])) {
			return nil, nil, fmt.Errorf("GenerateProofWithMerklePath tag no match with merkle tree error %s", err)
		}

		path, err := tree.GetMerklePath(uint64(challenge.Index))
		if err != nil {
			return nil, nil, fmt.Errorf("GenerateProofWithMerklePath GetMerklePath error %s", err)
		}
		pathResult = append(pathResult, path)
	}
	return proof, pathResult, nil
}

func (this *Pdp) VerifyProofWithMerklePath(version uint64, proofs []byte, fileIds []FileID, tags []Tag, challenges []Challenge, merklePath []*MerklePath, rootHashes [][]byte) error {
	// verify the merkle path for tags first
	if len(fileIds) != len(challenges) || len(fileIds) != len(tags) || len(fileIds) != len(challenges) || len(fileIds) != len(merklePath) || len(fileIds) != len(rootHashes) {
		return fmt.Errorf("VerifyProofWithMerklePath length of fileIds, tags, challenges and merklepath not the same")
	}

	if len(fileIds) == 0 {
		return fmt.Errorf("VerifyProofWithMerklePath length of fileIds is 0 ")
	}

	// verify merkle path for all tags
	for i, challenge := range challenges {
		path := merklePath[i]
		if len(path.Path) == 0 {
			return fmt.Errorf("VerifyProofWithMerklePath length of merkle path is 0 ")
		}

		err := VerifyMerklePath(path, uint64(challenge.Index), tags[i][:], rootHashes[i])
		if err != nil {
			return fmt.Errorf("VerifyProofWithMerklePath veriy merkle path error %s", err)
		}
	}

	if !this.VerifyProof(version, proofs, fileIds, tags, challenges) {
		return fmt.Errorf("VerifyProofWithMerklePath veriy proof failed")
	}
	return nil
}

// wrapper func to avoid the caller from the need to making the fileId slice for one file
func (this *Pdp) GenerateProofForFile(version uint64, blocks []Block, fileId FileID, challenges []Challenge) ([]byte, error) {
	return this.GenerateProof(version, blocks, fillFileIDs(fileId, len(blocks)), challenges)
}

func (this *Pdp) VerifyProofForFile(version uint64, proofs []byte, fileId FileID, tags []Tag, challenges []Challenge) bool {
	return this.VerifyProof(version, proofs, fillFileIDs(fileId, len(tags)), tags, challenges)
}

// wrapper func to avoid the caller from the need to making the fileId slice for one file
func (this *Pdp) GenerateProofWithMerklePathForFile(version uint64, blocks []Block, fileId FileID, tags []Tag, challenges []Challenge) ([]byte, []*MerklePath, error) {
	return this.GenerateProofWithMerklePath(version, blocks, fillFileIDs(fileId, len(blocks)), tags, challenges)
}

// wrapper func to avoid the caller from making the fileId and rootHash slice for one file
func (this *Pdp) VerifyProofWithMerklePathForFile(version uint64, proofs []byte, fileId FileID, tags []Tag, challenges []Challenge, merklePath []*MerklePath, rootHash []byte) error {
	return this.VerifyProofWithMerklePath(version, proofs, fillFileIDs(fileId, len(tags)), tags, challenges, merklePath, fillRootHashes(rootHash, len(tags)))
}

func checkBlockAndAddPadding(block Block) Block {
	if len(block) == 0 || len(block) > BLOCK_LENGTH {
		return nil
	}

	if len(block) == BLOCK_LENGTH {
		return block
	}

	tmp := make([]byte, BLOCK_LENGTH)
	copy(tmp, block)
	return tmp
}

func convertBlocks(blocks []Block) ([]bp.Block, error) {
	bpBlocks := make([]bp.Block, 0)
	for _, block := range blocks {
		padded := checkBlockAndAddPadding(block)
		if padded == nil {
			return nil, fmt.Errorf("block length exceed limit")
		}
		bpBlocks = append(bpBlocks, bp.Block{padded[:]})
	}
	return bpBlocks, nil
}

func convertChallenges(challenges []Challenge) []bp.Challenge {
	bpChallenges := make([]bp.Challenge, 0)
	for _, challenge := range challenges {
		bpChallenges = append(bpChallenges, bp.Challenge{uint32(challenge.Index), uint32(challenge.Rand)})
	}
	return bpChallenges
}

func convertTags(tags []Tag) [][32]byte {
	bpTags := make([][32]byte, 0)
	for _, tag := range tags {
		bpTags = append(bpTags, tag)
	}
	return bpTags
}

func convertFileIDs(fileIDs []FileID) [][32]byte {
	bpFileId := make([][32]byte, 0)

	for _, fileID := range fileIDs {
		bpFileId = append(bpFileId, fileID)
	}
	return bpFileId
}

func fillFileIDs(fileId FileID, count int) []FileID {
	fileIds := make([]FileID, 0)
	for i := 0; i < count; i++ {
		fileIds = append(fileIds, fileId)
	}
	return fileIds
}

func fillRootHashes(rootHash []byte, count int) [][]byte {
	rootHashes := make([][]byte, 0)
	for i := 0; i < count; i++ {
		data := make([]byte, len(rootHash))
		copy(data, rootHash)
		rootHashes = append(rootHashes, data)
	}
	return rootHashes
}
