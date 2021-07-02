package bulletproof_pdp_25519

import (
	"fmt"
	"github.com/magiconair/properties/assert"
	"io"
	"os"
	"testing"
	"time"
)

func TestInnerProductPDP_ProofVerify(t *testing.T) {
	var pdp PDPAlgo
	pdp = NewInnerProductPDP()

	blockstotal := GetBlocks("./test/bigger_than_256k_fake.txt")
	fileID, _ := GenerateFileID("./test/bigger_than_256k.txt")

	blockstotal_real := GetBlocks("./test/bigger_than_256k.txt")

	filelength := uint32(len(blockstotal))

	initstart := time.Now()
	tags := pdp.GenTag(blockstotal_real, fileID)

	fmt.Printf("Generate %v tags takes: %v \n", filelength, time.Now().Sub(initstart))

	chalength := 6

	challenges := make([]Challenge, chalength)
	for i := 0; i < chalength; i++ {
		challenges[i] = Challenge{uint32(i), uint32(i)}
	}

	blocks := make([]Block, len(challenges))
	for i, cha := range challenges {
		blocks[i] = blockstotal[cha.Index%filelength]
	}
	blocks_real := make([]Block, len(challenges))
	for i, cha := range challenges {
		blocks_real[i] = blockstotal_real[cha.Index%filelength]
	}

	verifyTags := make([][32]byte, chalength)
	for i, cha := range challenges {
		verifyTags[i] = tags[cha.Index%filelength]
	}

	fileIDList := make([][32]byte,chalength)
	for i := range fileIDList{
		fileIDList[i] = fileID
	}

	proof_real, _ := pdp.ProofGenerate(0, blocks_real, fileIDList, challenges)
	t0 := time.Now()
	proof, err := pdp.ProofGenerate(0, blocks, fileIDList, challenges)
	if err != nil {
		t.Fatal(err)
	}
	t1 := time.Now()
	result := pdp.ProofVerify(0, proof, fileIDList, verifyTags, challenges)
	t2 := time.Now()
	result_real := pdp.ProofVerify(0, proof_real, fileIDList, verifyTags, challenges)

	fmt.Println("proof generate takes: ", t1.Sub(t0))
	fmt.Println("proof verify takes: ", t2.Sub(t1))
	assert.Equal(t,result,false)
	assert.Equal(t,result_real,true)
}

func TestInnerProductPDP_DiffFiles(t *testing.T) {
	var pdp PDPAlgo
	pdp = NewInnerProductPDP()

	fileID, _ := GenerateFileID("./test/bigger_than_256k.txt")
	fileID2,_ := GenerateFileID("./test/bigger_than_256k_2.txt")

	blockstotal:= GetBlocks("./test/bigger_than_256k.txt")
	blockstotal2:= GetBlocks("./test/bigger_than_256k_2.txt")

	tags := pdp.GenTag(blockstotal, fileID)
	tags2 := pdp.GenTag(blockstotal2, fileID2)

	chalength := 6

	challenges := make([]Challenge, chalength)
	for i := 0; i < chalength; i++ {
		challenges[i] = Challenge{uint32(i), uint32(i)}
	}

	blocks := []Block{blockstotal[0],blockstotal[1],blockstotal[2],blockstotal2[0],blockstotal2[1],blockstotal2[2]}

	verifyTags := [][32]byte{tags[0],tags[1],tags[2],tags2[0],tags2[1],tags2[2]}

	fileIDList := [][32]byte{fileID,fileID,fileID,fileID2,fileID2,fileID2}

	t0 := time.Now()
	proof, err := pdp.ProofGenerate(0, blocks, fileIDList, challenges)
	if err != nil {
		t.Fatal(err)
	}
	t1 := time.Now()
	result := pdp.ProofVerify(0, proof, fileIDList, verifyTags, challenges)
	t2 := time.Now()

	fmt.Println("proof generate takes: ", t1.Sub(t0))
	fmt.Println("proof verify takes: ", t2.Sub(t1))
	assert.Equal(t,result,true)
}

func GetBlocks(path string) []Block {
	fi, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer fi.Close()

	chunks := make([]Block, 0)

	for {
		buf := make([]byte, blocksize)
		n, err := fi.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if 0 == n {
			break
		}

		chunks = append(chunks, Block{buf})
	}

	return chunks
}

