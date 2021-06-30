package pdp

import (
	"crypto/rand"
	"fmt"
	rand2 "math/rand"
	"sort"
	"testing"
)

type FileOption struct {
	fileId       FileID
	challengeNum int
	blockNum     int
	allTag       bool
}

// pdp test without merkle path verification for a file
func TestPDP(t *testing.T) {
	files := []FileOption{
		{
			fileId:       generateFileId(),
			challengeNum: 2,
			blockNum:     100,
			allTag:       false,
		},
		{
			fileId:       generateFileId(),
			challengeNum: 3,
			blockNum:     200,
			allTag:       false,
		},
		{
			fileId:       generateFileId(),
			challengeNum: 1,
			blockNum:     20,
			allTag:       false,
		},
	}
	for _, file := range files {
		err := doPDPTest(file)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func doPDPTest(file FileOption) error {
	p := NewPdp(0)

	fileData, err := generateFileData(file, p)
	if err != nil {
		return err
	}

	proof, err := p.GenerateProofForFile(0, fileData.blocksForPdp, file.fileId, fileData.challenges)
	if err != nil {
		return fmt.Errorf("generateProof error %s", err)
	}

	// use another instance for verification
	p2 := NewPdp(0)
	ok := p2.VerifyProofForFile(0, proof, file.fileId, fileData.tagsForPdp, fileData.challenges)
	if !ok {
		return fmt.Errorf("verify proof nok")
	}
	return nil
}

// pdp test with merkle path verification for a file
func TestPDPMerklePath(t *testing.T) {
	files := []FileOption{
		{
			fileId:       generateFileId(),
			challengeNum: 2,
			blockNum:     100,
			allTag:       true,
		},
		{
			fileId:       generateFileId(),
			challengeNum: 3,
			blockNum:     200,
			allTag:       true,
		},
		{
			fileId:       generateFileId(),
			challengeNum: 1,
			blockNum:     20,
			allTag:       true,
		},
	}
	for _, file := range files {
		err := doPDPMerklePathTest(file)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func doPDPMerklePathTest(file FileOption) error {
	p := NewPdp(0)

	fileData, err := generateFileData(file, p)
	if err != nil {
		return err
	}

	nodes := make([]*MerkleNode, 0)
	for i, tag := range fileData.tags {
		node := InitNodeWithData(tag[:], uint64(i))
		nodes = append(nodes, node)
	}

	err = p.InitMerkleTreeForFile(file.fileId, nodes)
	if err != nil {
		return err
	}

	rootHash, err := p.GetRootHashForFile(file.fileId)
	if err != nil {
		return err
	}

	fileId := file.fileId
	blocks := fileData.blocksForPdp
	tags := fileData.tagsForPdp
	challenges := fileData.challenges

	proof, mpath, err := p.GenerateProofWithMerklePathForFile(0, blocks, fileId, tags, challenges)
	if err != nil {
		return fmt.Errorf("generateProof error %s", err)
	}

	// use another instance for verification
	p2 := NewPdp(0)
	err = p2.VerifyProofWithMerklePathForFile(0, proof, fileId, tags, challenges, mpath, rootHash)
	if err != nil {
		return fmt.Errorf("VerifyProofWithMerklePath error %s", err)
	}
	return nil
}

// pdp test without merkle path verification for files
func TestMultiPDP(t *testing.T) {
	files := []FileOption{
		{
			fileId:       generateFileId(),
			challengeNum: 2,
			blockNum:     100,
			allTag:       false,
		},
		{
			fileId:       generateFileId(),
			challengeNum: 3,
			blockNum:     200,
			allTag:       false,
		},
		{
			fileId:       generateFileId(),
			challengeNum: 1,
			blockNum:     20,
			allTag:       false,
		},
	}

	err := doMultiPDPTest(files)
	if err != nil {
		t.Fatal(err)
	}
}

func doMultiPDPTest(files []FileOption) error {
	p := NewPdp(0)

	testData := make([]*FileData, 0)
	for _, option := range files {
		fileData, err := generateFileData(option, p)
		if err != nil {
			return err
		}
		testData = append(testData, fileData)
	}

	fileIds := make([]FileID, 0)
	tagsForPdp := make([]Tag, 0)
	blocksForPdp := make([]Block, 0)
	challenges := make([]Challenge, 0)

	for _, t := range testData {
		fileIds = append(fileIds, fillFileIDs(t.fileId, len(t.tagsForPdp))...)
		tagsForPdp = append(tagsForPdp, t.tagsForPdp...)
		blocksForPdp = append(blocksForPdp, t.blocksForPdp...)
		challenges = append(challenges, t.challenges...)

	}

	proof, err := p.GenerateProof(0, blocksForPdp, fileIds, challenges)
	if err != nil {
		return fmt.Errorf("generateProof error %s", err)
	}

	// use another instance for verification
	p2 := NewPdp(0)
	ok := p2.VerifyProof(0, proof, fileIds, tagsForPdp, challenges)
	if !ok {
		return fmt.Errorf("verify proof nok")
	}
	return nil
}

// pdp test with merkle path verification for files
func TestMultiPDPMerklePath(t *testing.T) {
	files := []FileOption{
		{
			fileId:       generateFileId(),
			challengeNum: 2,
			blockNum:     100,
			allTag:       true,
		},
		{
			fileId:       generateFileId(),
			challengeNum: 3,
			blockNum:     200,
			allTag:       true,
		},
		{
			fileId:       generateFileId(),
			challengeNum: 1,
			blockNum:     20,
			allTag:       true,
		},
	}
	err := doMultiPDPMerklePathTest(files)
	if err != nil {
		t.Fatal(err)
	}
}

func doMultiPDPMerklePathTest(files []FileOption) error {
	p := NewPdp(0)

	testData := make([]*FileData, 0)
	for _, option := range files {
		fileData, err := generateFileData(option, p)
		if err != nil {
			return err
		}

		nodes := make([]*MerkleNode, 0)
		for i, tag := range fileData.tags {
			node := InitNodeWithData(tag[:], uint64(i))
			nodes = append(nodes, node)
		}

		err = p.InitMerkleTreeForFile(option.fileId, nodes)
		if err != nil {
			return err
		}

		testData = append(testData, fileData)
	}

	fileIds := make([]FileID, 0)
	tagsForPdp := make([]Tag, 0)
	blocksForPdp := make([]Block, 0)
	challenges := make([]Challenge, 0)
	rootHashes := make([][]byte, 0)

	for _, t := range testData {
		fileIds = append(fileIds, fillFileIDs(t.fileId, len(t.tagsForPdp))...)
		tagsForPdp = append(tagsForPdp, t.tagsForPdp...)
		blocksForPdp = append(blocksForPdp, t.blocksForPdp...)
		challenges = append(challenges, t.challenges...)

		rootHash, err := p.GetRootHashForFile(t.fileId)
		if err != nil {
			return err
		}
		rootHashes = append(rootHashes, fillRootHashes(rootHash, len(t.tagsForPdp))...)
	}

	proof, mpath, err := p.GenerateProofWithMerklePath(0, blocksForPdp, fileIds, tagsForPdp, challenges)
	if err != nil {
		return fmt.Errorf("generateProof error %s", err)
	}

	// use another instance for verification
	p2 := NewPdp(0)
	err = p2.VerifyProofWithMerklePath(0, proof, fileIds, tagsForPdp, challenges, mpath, rootHashes)
	if err != nil {
		return fmt.Errorf("VerifyProofWithMerklePath error %s", err)
	}
	return nil
}

type FileData struct {
	fileId       FileID
	blocks       []Block
	tags         []Tag
	blocksForPdp []Block
	tagsForPdp   []Tag
	challenges   []Challenge
}

func generateFileData(file FileOption, p *Pdp) (*FileData, error) {
	blocks := make([]Block, 0)
	tags := make([]Tag, 0)

	blocksForPdp := make([]Block, 0)
	tagsForPdp := make([]Tag, 0)

	challenges, err := generateChallenges(file.blockNum, file.challengeNum)
	if err != nil {
		return nil, err
	}

	for i := 0; i < file.blockNum; i++ {
		data := make([]byte, BLOCK_LENGTH)
		rand.Read(data)

		blocks = append(blocks, data)

		var tag []Tag
		var err error

		if isChallenged(challenges, uint32(i)) {
			blocksForPdp = append(blocksForPdp, data)
			tag, err = p.GenerateTag([]Block{data}, file.fileId)
			if err != nil {
				return nil, fmt.Errorf("GenerateTag error %s", err)
			}

			tagsForPdp = append(tagsForPdp, tag...)
		} else {
			// also generate tag if not challenged but all tag needed for merkle path calculation
			if file.allTag {
				tag, err = p.GenerateTag([]Block{data}, file.fileId)
				if err != nil {
					return nil, fmt.Errorf("GenerateTag error %s", err)
				}
			}
		}

		tags = append(tags, tag...)
	}

	return &FileData{
		fileId:       file.fileId,
		blocks:       blocks,
		tags:         tags,
		blocksForPdp: blocksForPdp,
		tagsForPdp:   tagsForPdp,
		challenges:   challenges,
	}, nil
}

func generateFileId() FileID {
	var fileId FileID
	rand.Read(fileId[:])
	return fileId
}

func isChallenged(challenges []Challenge, index uint32) bool {
	for _, challenge := range challenges {
		if challenge.Index == index {
			return true
		}
	}
	return false
}
func generateChallenges(blockNum int, challengeNum int) ([]Challenge, error) {
	challenges := make([]Challenge, 0)
	indexMap := make(map[uint32]struct{})
	maxRetry := 3

	for i := 0; i < challengeNum; i++ {
		ok := false
		index := uint32(rand2.Int31n(int32(blockNum)))
		for retry := 0; retry < maxRetry; retry++ {
			if _, exist := indexMap[index]; exist {
				continue
			}
			ok = true
			indexMap[index] = struct{}{}
			break
		}

		if !ok {
			return nil, fmt.Errorf("generate challenge error")
		}

		challenge := Challenge{
			Index: index,
			Rand:  rand2.Uint32(),
		}
		challenges = append(challenges, challenge)
	}

	// sort the challenges for merkle path calculation
	sort.SliceStable(challenges, func(i, j int) bool {
		return challenges[i].Index < challenges[j].Index
	})
	return challenges, nil
}
