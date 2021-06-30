package pdp

type Block []byte
type Tag [TAG_LENGTH]byte
type FileID [FILEID_LENGTH]byte

const BLOCK_LENGTH = 256 * 1024
const TAG_LENGTH = 32
const FILEID_LENGTH = 32

type Challenge struct {
	Index uint32
	Rand  uint32
}
