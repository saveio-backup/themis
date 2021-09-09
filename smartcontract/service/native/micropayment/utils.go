package micropayment

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"math"
	"strconv"
	"strings"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/utils"
	"github.com/saveio/themis/vm/wasmvm/util"
)

const (
	MP_INIT                          = "MPInit"
	MP_OPEN_CHANNEL                  = "OpenChannel"
	MP_SET_TOTALDEPOSIT              = "SetTotalDeposit"
	MP_SET_TOTALWITHDRAW             = "SetTotalWithdraw"
	MP_COOPERATIVESETTLE             = "CooperativeSettle"
	MP_CLOSE_CHANNEL                 = "CloseChannel"
	MP_UNLOCK                        = "Unlock"
	MP_SECRET_REG                    = "RegisterSecret"
	MP_SECRET_REG_BATCH              = "RegisterSecretBatch"
	MP_GET_SECRET_REVEAL_BLOCKHEIGHT = "GetSecretRevealBlockHeight"
	MP_UPDATE_NONCLOSING_BPF         = "UpdateNonClosingBalanceProof"
	MP_SETTLE_CHANNEL                = "SettleChannel"
	MP_GET_CHANNELINFO               = "GetChannelInfo"
	MP_GET_ALL_OPEN_CHANNELS         = "GetAllOpenChannels"
	MP_GET_CHANNELCOUNTER            = "GetChannelCounter"
	MP_GET_CHANNEL_PARTICIPANTINFO   = "GetChannelParticipantInfo"
	MP_GET_CHANNELID                 = "GetChannelIdentifier"
	MP_GET_NODE_PUBKEY               = "GetNodePubKey"
	MP_SET_NODE_PUBKEY               = "SetNodePubKey"
	MP_FAST_TRANSFER                 = "FastTransfer"
	//MP_SETTLEMENT_TRANSFER = "SettlementTransfer"
)

func UpdateBalanceProofData(chanInfo *ChannelInfo, participant common.Address, nonce uint64, balanceHash []byte) error {
	updateParticipant := WhoIsParticipant(participant, chanInfo.Participant1.WalletAddr,
		&chanInfo.Participant1, &chanInfo.Participant2)
	if nonce <= updateParticipant.Nonce {
		log.Error("[UpdateBalanceProofData] input nonce lt updateParticipantNonce")
		return errors.NewErr("[UpdateBalanceProofData] input nonce lt updateParticipantNonce")
	}

	updateParticipant.Nonce = nonce
	updateParticipant.BalanceHash = balanceHash
	return nil
}

func GetChannelAvailableDeposit(participant1, participant2 Participant) uint64 {
	totalAvailableDeposit := participant1.Deposit + participant2.Deposit -
		participant1.WithDrawAmount - participant2.WithDrawAmount
	return totalAvailableDeposit
}

func WithDrawMessageBundleHash(channelID uint64, participant common.Address, totalWithdraw uint64) [32]byte {
	messageType := MessageType{}
	messageType.TypeId = WITHDRAW
	messageBundle := bytes.Join([][]byte{
		[]byte(SIGNATURE_PREFIX),
		[]byte(WITHDRAW_MESSAGE_LENGTH),
		util.Int64ToBytes(messageType.TypeId),
		util.Int64ToBytes(channelID),
		participant[:],
		util.Int64ToBytes(totalWithdraw),
	}, []byte{})

	return sha256.Sum256(messageBundle)
}

//==balance proof
func ClosedMessageBundleHash(channelID uint64, balanceHash []byte, nonce uint64, additionalHash []byte) [32]byte {
	messageType := MessageType{}
	messageType.TypeId = BalanceProof
	messageBundle := bytes.Join([][]byte{
		[]byte(SIGNATURE_PREFIX),              //"\x19Ontology Signed Message:\n"
		[]byte(CLOSE_MESSAGE_LENGTH),          //"212"
		util.Int64ToBytes(messageType.TypeId), // 1
		util.Int64ToBytes(channelID),
		balanceHash,
		util.Int64ToBytes(nonce),
		additionalHash,
	}, []byte{})

	return sha256.Sum256(messageBundle)
}

func CoSettleMessageBundleHash(channelID uint64, participant1 common.Address, participant1Balance uint64,
	participant2 common.Address, participant2Balance uint64) [32]byte {
	messageType := MessageType{}
	messageType.TypeId = COOPERATIVESETTLE
	messageBundle := bytes.Join([][]byte{
		[]byte(SIGNATURE_PREFIX),
		[]byte(COSETTLE_MESSAGE_LENGTH),       //220
		util.Int64ToBytes(messageType.TypeId), //4
		util.Int64ToBytes(channelID),
		participant1[:],
		util.Int64ToBytes(participant1Balance),
		participant2[:],
		util.Int64ToBytes(participant2Balance),
	}, []byte{})

	return sha256.Sum256(messageBundle)
}

//NonClose sig messageHash
func BalanceProofUpdateMessageBundleHash(channelID uint64, balanceHash []byte, nonce uint64,
	additionalHash []byte, closeSignature []byte) [32]byte {
	messageType := MessageType{}
	messageType.TypeId = BalanceProofUpdate
	messageBundle := bytes.Join([][]byte{
		[]byte(SIGNATURE_PREFIX),
		[]byte(BALANCEPROOF_UPDATE_MESSAGE_LENGTH),
		util.Int64ToBytes(messageType.TypeId),
		util.Int64ToBytes(channelID),
		balanceHash,
		util.Int64ToBytes(nonce),
		additionalHash,
		closeSignature,
	}, []byte{})

	return sha256.Sum256(messageBundle)
}

func GetChanInfoFromDB(native *native.NativeService, chanId uint64) (*ChannelInfo, error) {
	chIdBf := new(bytes.Buffer)
	utils.WriteVarUint(chIdBf, chanId)
	chInfoItem, err := utils.GetStorageItem(native, chIdBf.Bytes())
	if err != nil {
		return nil, errors.NewErr("[MPay Contract][GetChanInfoFromDB] Get chInfoItem from ChannelID error!")
	}

	if chInfoItem == nil {
		return nil, errors.NewErr("[MPay Contract][GetChanInfoFromDB] chInfoItem is not found!")
	}

	var chanInfo ChannelInfo
	chanInfoSource := common.NewZeroCopySource(chInfoItem.Value)
	err = chanInfo.Deserialization(chanInfoSource)
	if err != nil {
		return nil, errors.NewErr("[MPay Contract][GetChanInfoFromDB] chInfo Deserialization error!")
	}
	return &chanInfo, nil
}

func DeleteChanInfoFromDB(native *native.NativeService, chanId uint64) {
	chIdBf := new(bytes.Buffer)
	utils.WriteVarUint(chIdBf, chanId)
	utils.DelStorageItem(native, chIdBf.Bytes())
}

func PutChannelInfoToDB(native *native.NativeService, channel ChannelInfo) error {
	chInfoBf := new(bytes.Buffer)
	if err := channel.Serialize(chInfoBf); err != nil {
		return errors.NewErr("[MPay Contract][PutChannelInfoToDB] channelInfo serialize error!")
	}

	chIdBf := new(bytes.Buffer)
	utils.WriteVarUint(chIdBf, channel.ChannelID)
	utils.PutBytes(native, chIdBf.Bytes(), chInfoBf.Bytes())
	return nil
}

func getChannelIDfromKey(native *native.NativeService, key [32]byte) uint64 {
	chIdItem, err := utils.GetStorageItem(native, key[:])
	if err != nil {
		return 0
	}

	if chIdItem == nil {
		return 0
	}

	reader := bytes.NewReader(chIdItem.Value)
	chID, _ := utils.ReadVarUint(reader)
	return chID
}

func GetChannelID(native *native.NativeService, participant1, participant2 common.Address) uint64 {
	key := GetParticipantHash(participant1, participant2)
	chIdItem, err := utils.GetStorageItem(native, key[:])
	if err != nil {
		return 0
	}

	if chIdItem == nil {
		return 0
	}

	reader := bytes.NewReader(chIdItem.Value)
	chID, _ := utils.ReadVarUint(reader)
	return chID
}

func GetParticipantHash(participant1, participant2 common.Address) [32]byte {
	if string(participant1[:]) < string(participant2[:]) {
		return sha256.Sum256(append(participant1[:], participant2[:]...))
	} else {
		return sha256.Sum256(append(participant2[:], participant1[:]...))
	}
}

func GenPubKeyKey(contract common.Address, walletAddr common.Address) []byte {
	prefix := []byte("PubKeyKey")
	key := append(contract[:], prefix...)
	key = append(contract[:], walletAddr[:]...)
	return key
}

func ConvertByteSliceToString(b []byte) string {
	s := make([]string, len(b))
	for i := range b {
		s[i] = strconv.Itoa(int(b[i]))
	}
	return strings.Join(s, ",")
}

func WhoIsParticipant(participantWalletAddr, chP1WalletAddr common.Address, chP1, chP2 *Participant) *Participant {
	if participantWalletAddr == chP1WalletAddr {
		return chP1
	} else {
		return chP2
	}
}

func verifyBalanceHashData(participant Participant, transferredAmount, lockedAmount uint64, locksRoot []byte) bool {
	if isEmptyHash(participant.BalanceHash) && transferredAmount == 0 && lockedAmount == 0 && isEmptyHash(locksRoot) {
		return true
	}

	balanceHash := sha256.Sum256(bytes.Join(
		[][]byte{util.Int64ToBytes(transferredAmount),
			util.Int64ToBytes(lockedAmount), locksRoot}, []byte{}))
	return byteSliceEqual(participant.BalanceHash, balanceHash[:])
}

func isEmptyHash(hash []byte) bool {
	var emptyHash [32]byte

	if len(hash) == 0 {
		return true
	}
	return byteSliceEqual(hash, emptyHash[:])
}

func byteSliceEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) {
		return false
	}

	b = b[:len(a)]
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func getSettleTransferAmounts(p1 Participant, p1TransferredAmount, p1LockedAmount uint64,
	p2 Participant, p2TransferredAmount, p2LockedAmount uint64) (uint64, uint64, uint64, uint64, error) {
	var participant1Settlement, participant2Settlement SettlementData
	var participant1Amount, participant2Amount, totalAvailableDeposit uint64
	var err error

	participant1Settlement.Deposit = p1.Deposit
	participant1Settlement.Withdrawn = p1.WithDrawAmount
	participant1Settlement.Transferred = p1TransferredAmount
	participant1Settlement.Locked = p1LockedAmount

	participant2Settlement.Deposit = p2.Deposit
	participant2Settlement.Withdrawn = p2.WithDrawAmount
	participant2Settlement.Transferred = p2TransferredAmount
	participant2Settlement.Locked = p2LockedAmount

	totalAvailableDeposit = GetChannelAvailableDeposit(p1, p2)
	participant1Amount, err = getMaxPossibleReceivableAmount(participant1Settlement, participant2Settlement)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	participant1Amount = minAmount(participant1Amount, totalAvailableDeposit)
	participant2Amount = totalAvailableDeposit - participant1Amount
	participant1Amount, p2LockedAmount = failSafeSubstract(participant1Amount, p2LockedAmount)
	participant2Amount, p1LockedAmount = failSafeSubstract(participant2Amount, p1LockedAmount)

	if participant1Amount > totalAvailableDeposit {
		log.Error("[getSettleTransferAmounts] p1Amount >TAD")
		return 0, 0, 0, 0, errors.NewErr("[getSettleTransferAmounts] p1Amount >TAD")
	}

	if participant2Amount > totalAvailableDeposit {
		log.Error("[getSettleTransferAmounts] p2Amount >TAD")
		return 0, 0, 0, 0, errors.NewErr("[getSettleTransferAmounts] p2Amount >TAD")
	}
	if totalAvailableDeposit != (participant1Amount + participant2Amount + p1LockedAmount + p2LockedAmount) {
		log.Error("[getSettleTransferAmounts]TAD ne P1+P2")
		return 0, 0, 0, 0, errors.NewErr("[getSettleTransferAmounts]TAD ne P1+P2")
	}
	return participant1Amount, participant2Amount, p1LockedAmount, p2LockedAmount, nil

}

//@notice: assume p2 transferred more than p1! Notice the param order!
func getMaxPossibleReceivableAmount(participant1Settlement, participant2Settlement SettlementData) (uint64, error) {
	var participant1MaxTransferred uint64
	var participant2MaxTransferred uint64
	var participant1NetMaxReceived uint64
	var participant1MaxAmount uint64

	participant1MaxTransferred = failSafeAddUint64(
		participant1Settlement.Transferred,
		participant1Settlement.Locked,
	)
	participant2MaxTransferred = failSafeAddUint64(
		participant2Settlement.Transferred,
		participant2Settlement.Locked,
	)

	//Require(participant2MaxTransferred>=participant1MaxTransferred,"")
	if participant1MaxTransferred < participant1Settlement.Transferred {
		log.Error("P1Transferred ge maxTransferred")
		return 0, errors.NewErr("P1Transferred ge maxTransferred")
	}

	if participant2MaxTransferred < participant2Settlement.Transferred {
		log.Error("P2Transferred ge maxTransferred")
		return 0, errors.NewErr("P2Transferred ge maxTransferred")
	}

	participant1NetMaxReceived = participant2MaxTransferred - participant1MaxTransferred
	participant1MaxAmount = failSafeAddUint64(participant1NetMaxReceived, participant1Settlement.Deposit)

	participant1MaxAmount = participant1MaxAmount - participant1Settlement.Withdrawn

	return participant1MaxAmount, nil
}

func storeUnlockData(native *native.NativeService, chanID uint64, participant, partner common.Address,
	lockedAmount uint64, locksRoot []byte) error {
	if lockedAmount == 0 || locksRoot == nil {
		return nil
	}

	key, err := getUnlockId(chanID, participant, partner)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	var unlockData UnlockDataInfo
	unlockData.LocksRoot = locksRoot
	unlockData.LockedAmount = lockedAmount
	bf := new(bytes.Buffer)
	unlockData.Serialize(bf)
	utils.PutBytes(native, key, bf.Bytes())
	return nil
}

func getUnlockId(chanID uint64, participant, partner common.Address) ([]byte, error) {
	if participant.ToBase58() == partner.ToBase58() {
		log.Error("[getUnlockId]participant==partner error!")
		return nil, errors.NewErr("[getUnlockId]participant==partner error!")
	}

	unlockID := sha256.Sum256(bytes.Join(
		[][]byte{util.Int64ToBytes(chanID),
			participant[:], partner[:]}, []byte{}))
	return unlockID[:], nil
}

func getMerkleRootAndUnlockedAmount(native *native.NativeService, merkleTreeLeaves []byte) ([]byte, uint64, error) {
	var totalUnlockedAmount uint64
	var unlockedAmount uint64
	var lockHash []byte

	// each merkle_tree lock component has this form:
	// (locked_amount || expiration_block || secrethash) = 8 + 8 + 32
	length := len(merkleTreeLeaves)
	if length%48 != 0 {
		log.Error("[getMerkleRootAndUnlockedAmount] invalid merkle tree leaves length!")
		return nil, 0, errors.NewErr("[getMerkleRootAndUnlockedAmount] invalid merkle tree leaves length!")
	}

	merkleLayer := make([][]byte, length/48+1)

	// why i start from 32? may need to change to 0
	for i := 0; i < length; i += 48 {
		lockHash, unlockedAmount = getLockDataFromMerkleTree(native, merkleTreeLeaves, uint64(i))
		totalUnlockedAmount += unlockedAmount
		merkleLayer[i/48] = lockHash
	}

	length = length / 48

	for {
		if length > 1 {
			if length%2 != 0 {
				merkleLayer[length] = merkleLayer[length-1]
				length++
			}

			var i int
			for i = 0; i < length-1; i += 2 {
				result := bytes.Compare(merkleLayer[i], merkleLayer[i+1])

				if result == 0 {
					lockHash = merkleLayer[i]
				} else if result < 0 {
					sum := sha256.Sum256(bytes.Join(
						[][]byte{merkleLayer[i], merkleLayer[i+1]}, []byte{}))
					lockHash = sum[:]
				} else {
					sum := sha256.Sum256(bytes.Join(
						[][]byte{merkleLayer[i+1], merkleLayer[i]}, []byte{}))
					lockHash = sum[:]
				}

				merkleLayer[i/2] = lockHash
			}
			length = i / 2
		} else {
			break
		}
	}

	return merkleLayer[0], totalUnlockedAmount, nil
}

func getLockDataFromMerkleTree(native *native.NativeService, merkleTreeLeaves []byte, offset uint64) ([]byte, uint64) {
	var expirationBlock uint64
	var lockedAmount uint64
	var secretHash []byte
	var revealBlock uint64

	if len(merkleTreeLeaves) <= int(offset) {
		return nil, 0
	}

	expirationBlock = bytesToUint64(merkleTreeLeaves[offset : offset+8])
	lockedAmount = bytesToUint64(merkleTreeLeaves[offset+8 : offset+16])
	secretHash = merkleTreeLeaves[offset+16 : offset+48]

	// calculae the lockhash for computing the merkle root
	sum := sha256.Sum256(merkleTreeLeaves[offset : offset+48])
	lockHash := sum[:]

	// check if the lock's secret was revealed and the secret is revealed before expire
	revealBlockValue, err := getSecretRevealBlockHeight(native, secretHash)
	if err != nil || revealBlockValue == nil {
		revealBlock = 0
	} else {
		reader := bytes.NewReader(revealBlockValue)
		revealBlock, _ = utils.ReadVarUint(reader)
	}

	if revealBlock == 0 || expirationBlock <= revealBlock {
		lockedAmount = 0
	}

	return lockHash, lockedAmount
}

func bytesToUint64(data []byte) uint64 {
	var n uint64
	bytesBuffer := bytes.NewBuffer(data)
	binary.Read(bytesBuffer, binary.LittleEndian, &n)
	return n
}

func failSafeAddUint64(left, right uint64) uint64 {
	sum := left + right
	if sum >= left {
		return sum
	}
	return math.MaxUint64
}

func minAmount(left, right uint64) uint64 {
	if left < right {
		return left
	}
	return right
}

func maxAmount(left, right uint64) uint64 {
	if left > right {
		return left
	}
	return right
}

func failSafeSubstract(left, right uint64) (uint64, uint64) {
	if left > right {
		return left - right, right
	}
	return 0, left

}
