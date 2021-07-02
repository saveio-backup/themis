package bulletproof_pdp_25519

/*
Block 数据块大小为固定参数blocksize(256*1024)字节。当文件长度无法被256KB等分时，应填充直到最后一个Block为256KB。

GenTag:生成输入数据块的标签
	params：
		blocks：数据块集合，数据块来自同一个文件对应相同的fileID
		fileID：文件唯一标识，为公开的随机字符串，同文件一一对应。可根据文件名或文件根哈希等参数生成。（32字节）
	return：数据块的标签序列化后的集合，每个标签序列化为[32]byte

ProofGenerate：存储节点生成证明
	params:
		pdpversion:pdp算法版本号
		blocks：被抽到挑战的数据块集合
		fileID：文件唯一标识
		challenges：挑战集合，顺序同blocks对应。
	return：证明序列

ProofVerify:数据完整性验证
	params：
		pdpversion:pdp算法版本号
		proofs：证明序列
		fileID：文件唯一标识
		blockTags: 被挑战数据块的标签序列化后的集合，每个标签序列化为[32]byte，顺序需同proofs对应。
		challenges：挑战集合，顺序同proofs对应。
	return：验证结果
*/
type PDPAlgo interface {
	GenTag(block []Block, fileID [32]byte) (blockTags [][32]byte)
	ProofGenerate(pdpversion uint64, blocks []Block, fileID [][32]byte, challenges []Challenge) ([]byte, error)
	ProofVerify(pdpversion uint64, proofs []byte, fileID [][32]byte, blockTags [][32]byte, challenges []Challenge) bool
}
