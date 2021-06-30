package micropayment

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/core/types"
	"github.com/saveio/themis/crypto/keypair"
	"github.com/saveio/themis/crypto/signature"
	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/usdt"
	"github.com/saveio/themis/smartcontract/service/native/utils"
	"github.com/saveio/themis/vm/wasmvm/util"
)

const (
	None = iota
	BalanceProof
	BalanceProofUpdate
	WITHDRAW
	COOPERATIVESETTLE
	SIGNATURE_PREFIX                   = "\x19Ontology Signed Message:\n"
	WITHDRAW_MESSAGE_LENGTH            = "168"
	COSETTLE_MESSAGE_LENGTH            = "220"
	CLOSE_MESSAGE_LENGTH               = "212"
	BALANCEPROOF_UPDATE_MESSAGE_LENGTH = "277"
	SECRET_LENGTH                      = 32
)

type MessageType struct {
	TypeId uint64
}

type SettlementData struct {
	Deposit     uint64
	Withdrawn   uint64
	Transferred uint64
	Locked      uint64
}

func RegisterSecret(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	secret, err := utils.DecodeBytes(source)
	if err != nil {
		log.Error("[MPay Contract][RegisterSecret] Secret decode error!")
		return utils.BYTE_FALSE, errors.NewErr("[MPay Contract][RegisterSecret] Secret decode error!")
	}

	err = registerSecret(native, secret)
	if err != nil {
		log.Error(err.Error())
		return utils.BYTE_FALSE, err
	}
	return utils.BYTE_TRUE, nil
}

// internal call
func registerSecret(native *native.NativeService, secret []byte) error {
	if len(secret) != SECRET_LENGTH {
		return errors.NewErr("[MPay Contract][RegisterSecret] incorrect secret length")
	}

	isValid := false
	for _, val := range secret {
		if val != 0 {
			isValid = true
			break
		}
	}

	if !isValid {
		return errors.NewErr("[MPay Contract][RegisterSecret] secret all zero")
	}

	// check if stored already
	secretHash := sha256.Sum256(secret)
	blockHeight, err := utils.GetStorageItem(native, secretHash[:])
	if err == nil && blockHeight != nil {
		return errors.NewErr("[MPay Contract][RegisterSecret] secret already registered")
	}

	buff := new(bytes.Buffer)
	utils.WriteVarUint(buff, uint64(native.Height))
	utils.PutBytes(native, secretHash[:], buff.Bytes())

	var secretArray [SECRET_LENGTH]byte
	copy(secretArray[:], secret[:])

	SecretRevealedEvent(native, secretHash, secretArray, native.Height)
	return nil
}

func RegisterSecretBatch(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	secrets, err := utils.DecodeBytes(source)
	if err != nil {
		log.Error("[MPay Contract][RegisterSecretBatch] Secrets decode error!")
		return utils.BYTE_FALSE, errors.NewErr("[MPay Contract][RegisterSecretBatch] Secrets decode error!")
	}

	flag := utils.BYTE_TRUE
	for i := 0; i < len(secrets); i += SECRET_LENGTH {
		err = registerSecret(native, secrets[i:i+SECRET_LENGTH])
		if err != nil {
			log.Error(err.Error())
			flag = utils.BYTE_FALSE
		}
	}
	return flag, nil
}

func GetSecretRevealBlockHeight(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	secretHash, err := utils.DecodeBytes(source)
	if err != nil {
		log.Error("[MPay Contract] getSecretRevealBlockHeight secretHash deserialize error!")
		return utils.BYTE_FALSE, errors.NewErr("[MPay Contract] getSecretRevealBlockHeight secretHash deserialize error!")
	}

	height, err := getSecretRevealBlockHeight(native, secretHash)
	if err != nil {
		log.Errorf(err.Error())
		return utils.BYTE_FALSE, err
	}
	return height, nil
}

func getSecretRevealBlockHeight(native *native.NativeService, secretHash []byte) ([]byte, error) {
	blockHeight, err := utils.GetStorageItem(native, secretHash)
	if err != nil || blockHeight == nil {
		return utils.BYTE_FALSE, errors.NewErr("[MPay Contract] getSecretRevealBlockHeight GetStorageItem from secretHash error!")
	}
	return blockHeight.Value, nil
}

func Unlock(native *native.NativeService) ([]byte, error) {
	var unlockInfo UnlockInfo

	contract := native.ContextRef.CurrentContext().ContractAddress
	//unlock param deserialization
	source := common.NewZeroCopySource(native.Input)
	if err := unlockInfo.Deserialization(source); err != nil {
		log.Error("[Unlock] UnlockInfo deserialization error!")
		return utils.BYTE_FALSE, errors.NewErr("[Unlock] UnlockInfo deserialization error!")
	}

	//check channel identifier
	if unlockInfo.ChannelID == GetChannelID(native, unlockInfo.ParticipantAddress, unlockInfo.PartnerAddress) {
		log.Error("[Unlock] channel identifier should be deleted!")
		return utils.BYTE_FALSE, errors.NewErr("[Unlock] channel identifier should be deleted!")
	}

	//check channel is not existed
	_, err := GetChanInfoFromDB(native, unlockInfo.ChannelID)
	if err == nil {
		log.Error("[Unlock] GetChanInfoFromDB should return no channel info!")
		return utils.BYTE_FALSE, errors.NewErr("[Unlock] GetChanInfoFromDB should return no channel info!")
	}

	if len(unlockInfo.MerkleTreeLeaves) == 0 {
		log.Error("[Unlock] merkel tree leaves length should not be 0")
		return utils.BYTE_FALSE, errors.NewErr("[Unlock] merkel tree leaves length should not be 0")
	}

	//calculate the locksroot for the pending transfers and the amout of
	//tokens corresponding to the locke transfers with secrets revealed on chain
	computedLocksroot, unlockedAmount, err := getMerkleRootAndUnlockedAmount(native, unlockInfo.MerkleTreeLeaves)
	if err != nil {
		log.Error("[Unlock] get merkle root and unlocked amount error")
		return utils.BYTE_FALSE, errors.NewErr("[Unlock] get merkle root and unlocked amount error")
	}

	//The partner must have a non-empty locksroot on-chain that must be the same as the computed locksroot
	//Get the amount of tokens that have been left in the contract, to account for pending transfers `partner` -> `participant`
	unlockKey, err := getUnlockId(unlockInfo.ChannelID, unlockInfo.PartnerAddress, unlockInfo.ParticipantAddress)
	if err != nil {
		log.Error("[Unlock] get unlock id error")
		return utils.BYTE_FALSE, errors.NewErr("[Unlock] get unlock id error")
	}

	unlockItem, err := utils.GetStorageItem(native, unlockKey)
	if err != nil || unlockItem == nil {
		log.Error("[Unlock] get unlock data from db error")
		return utils.BYTE_FALSE, errors.NewErr("[Unlock] get unlock data from db error")
	}

	unlockData := new(UnlockDataInfo)
	unlockDataInfoSource := common.NewZeroCopySource(unlockItem.Value)
	err = unlockData.Deserialization(unlockDataInfoSource)
	if err != nil {
		log.Errorf("[Unlock] unlock data info deserialize error!")
		return utils.BYTE_FALSE, errors.NewErr("[Unlock] unlock data info deserialize error!")
	}

	lockedAmount := unlockData.LockedAmount

	if bytes.Compare(unlockData.LocksRoot, computedLocksroot) != 0 {
		log.Error("[Unlock] stored locksroot not same as computed locksroot")
		return utils.BYTE_FALSE, errors.NewErr("[Unlock] stored locksroot not same as computed locksroot")
	}

	if lockedAmount <= 0 {
		log.Error("[Unlock] stored locked amount no larger than 0")
		return utils.BYTE_FALSE, errors.NewErr("[Unlock] stored locked amount no larger than 0")
	}

	// Make sure we don't transfer more tokens than previously reserved in
	// the smart contract.
	unlockedAmount = minAmount(unlockedAmount, lockedAmount)

	// Transfer the rest of the tokens back to the partner
	returnedTokens := lockedAmount - unlockedAmount

	// Remove partner's unlock data
	utils.DelStorageItem(native, unlockKey)

	// Transfer the unlocked tokens to the participant. unlocked_amount can
	// be 0
	if unlockedAmount > 0 {
		err = appCallTransfer(native, utils.UsdtContractAddress, contract, unlockInfo.ParticipantAddress, unlockedAmount)
		if err != nil {
			log.Error("[Unlock] appCallTransfer to participant error!")
			return utils.BYTE_FALSE, errors.NewErr("[Unlock] appCallTransfer to participant error!")
		}
	}

	// Transfer the rest of the tokens back to the partner
	if returnedTokens > 0 {
		err = appCallTransfer(native, utils.UsdtContractAddress, contract, unlockInfo.PartnerAddress, returnedTokens)
		if err != nil {
			log.Error("[Unlock] appCallTransfer to participant error!")
			return utils.BYTE_FALSE, errors.NewErr("[Unlock] appCallTransfer to participant error!")
		}
	}

	var event unlockEvent
	event.channelIdentifier = unlockInfo.ChannelID
	event.participant = unlockInfo.ParticipantAddress
	event.partner = unlockInfo.PartnerAddress
	event.unlockedAmount = unlockedAmount
	event.returnedTokens = returnedTokens
	copy(event.computedLocksroot[:], computedLocksroot[0:32])
	UnlockEvent(native, event, []common.Address{event.participant, event.partner})

	return utils.BYTE_TRUE, nil
}

func GetChannelCounter(native *native.NativeService) ([]byte, error) {
	var channelCounter uint64
	channelCounterItem, err := utils.GetStorageItem(native, []byte("channelCounter"))
	if err != nil || channelCounterItem == nil {
		channelCounter = 100
		buff := new(bytes.Buffer)
		utils.WriteVarUint(buff, channelCounter)
		utils.PutBytes(native, []byte("channelCounter"), buff.Bytes())
	}

	if channelCounterItem != nil {
		reader := bytes.NewReader(channelCounterItem.Value)
		channelCounter, err = utils.ReadVarUint(reader)
		if err != nil {
			log.Error("[GetChannelCounter] channelCounter ReadVarUint error")
			return utils.BYTE_FALSE, errors.NewErr("[GetChannelCounter] channelCounter ReadVarUint error")
		}
	}
	return util.Int64ToBytes(channelCounter), nil
}

func GetNodePubKey(native *native.NativeService) ([]byte, error) {
	var nodePubKey NodePubKey
	source := common.NewZeroCopySource(native.Input)
	if err := nodePubKey.Deserialization(source); err != nil {
		log.Error("[GetNodePubKey] NodePubKey deserialization error!")
		return utils.BYTE_FALSE, errors.NewErr("[GetNodePubKey] NodePubKey deserialization error!")
	}

	key := GenPubKeyKey(utils.MicroPayContractAddress, nodePubKey.Participant)
	item, err := utils.GetStorageItem(native, key)
	if err != nil {
		log.Error("[GetNodePubKey] GetStorageItem error")
		return utils.BYTE_FALSE, errors.NewErr("[GetNodePubKey] GetStorageItem error")
	}
	if item != nil { //is not set
		nodePubKey.PublicKey = item.Value
	}
	bf := new(bytes.Buffer)
	err = nodePubKey.Serialize(bf)
	if err != nil {
		log.Error("[GetNodePubKey] NodePubKey Serialize error!")
		return utils.BYTE_FALSE, errors.NewErr("[GetNodePubKey] NodePubKey Serialize error!")
	}
	return bf.Bytes(), nil
}

//@notice: channelID is global monotonically increasing.
func OpenChannel(native *native.NativeService) ([]byte, error) {
	var openCh OpenChannelInfo
	var ch ChannelInfo
	var channelCounter uint64
	var channelIdentifier uint64

	//openCh param deserialization
	source := common.NewZeroCopySource(native.Input)
	if err := openCh.Deserialization(source); err != nil {
		log.Error("[OpenChannel] OpenChannelInfo deserialization error!")
		return utils.BYTE_FALSE, errors.NewErr("[OpenChannel] OpenChannelInfo deserialization error!")
	}

	if len(openCh.Participant1PubKey) != 0 {
		wallet1PubKey, err := keypair.DeserializePublicKey(openCh.Participant1PubKey)
		if err != nil {
			log.Error("[OpenChannel] DeserializePublicKey error")
			return utils.BYTE_FALSE, errors.NewErr("[OpenChannel] DeserializePublicKey error")
		}

		wallet1Addr := types.AddressFromPubKey(wallet1PubKey)
		if wallet1Addr != openCh.Participant1WalletAddr {
			log.Error("[OpenChannel] PublicKey is not match error")
			return utils.BYTE_FALSE, errors.NewErr("[OpenChannel] PublicKey is not match error")
		}

		key := GenPubKeyKey(utils.MicroPayContractAddress, wallet1Addr)
		utils.PutBytes(native, key, openCh.Participant1PubKey)
	}

	channelCounterValue, err := GetChannelCounter(native)
	if err != nil {
		log.Error("[OpenChannel] getChannelCounter error")
		return utils.BYTE_FALSE, errors.NewErr("[OpenChannel] getChannelCounter error")
	}

	channelCounter = bytesToUint64(channelCounterValue)

	channelCounter += 1
	channelIdentifier = channelCounter
	//Get chid from DB,verify if them had openChannelID.
	pairHash := GetParticipantHash(openCh.Participant1WalletAddr, openCh.Participant2WalletAddr)
	chid := getChannelIDfromKey(native, pairHash)
	if chid != 0 {
		log.Errorf("The channelID %d with %s - %s  have existed.", openCh.Participant1WalletAddr.ToBase58(),
			openCh.Participant2WalletAddr.ToBase58(), chid)
		return utils.BYTE_FALSE,
			fmt.Errorf("The channelID %d with %s - %s  have existed.", openCh.Participant1WalletAddr.ToBase58(),
				openCh.Participant2WalletAddr.ToBase58(), chid)
	}
	//participantsHash=>ChannelId
	chidbuff := new(bytes.Buffer)

	utils.WriteVarUint(chidbuff, channelIdentifier)
	utils.PutBytes(native, pairHash[:], chidbuff.Bytes())

	//ChannelId=>chaninfo
	ch.ChannelState = Opened
	ch.ChannelID = channelIdentifier
	ch.Participant1.WalletAddr = openCh.Participant1WalletAddr
	ch.Participant2.WalletAddr = openCh.Participant2WalletAddr
	ch.SettleBlockHeight = openCh.SettleBlockHeight

	chaninfobf := new(bytes.Buffer)
	if err = ch.Serialize(chaninfobf); err != nil {
		log.Error(("[MPay Contract][OpenChannel] channelInfo serialize error!"))
		return utils.BYTE_FALSE, errors.NewErr("[MPay Contract][OpenChannel] channelInfo serialize error!")
	}
	utils.PutBytes(native, chidbuff.Bytes(), chaninfobf.Bytes())

	//put new cc into DB
	ccbuff := new(bytes.Buffer)
	utils.WriteVarUint(ccbuff, channelCounter)
	utils.PutBytes(native, []byte("channelCounter"), ccbuff.Bytes())

	//EventNotify
	var chanOpen channelOpenedEvent
	chanOpen.channelIdentifier = channelIdentifier
	chanOpen.participant1 = ch.Participant1.WalletAddr
	chanOpen.participant2 = ch.Participant2.WalletAddr
	chanOpen.settleTimeout = ch.SettleBlockHeight
	ChannelOpenedEvent(native, chanOpen, []common.Address{chanOpen.participant1, chanOpen.participant2})

	return util.Int64ToBytes(channelIdentifier), nil
}

func SetTotalDeposit(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	var std SetTotalDepositInfo
	source := common.NewZeroCopySource(native.Input)
	err := std.Deserialization(source)
	if err != nil {
		log.Error("[SetTotalDeposit] ChannelInfo deserialization error!")
		return utils.BYTE_FALSE, errors.NewErr("[SetTotalDeposit] ChannelInfo deserialization error!")
	}

	if std.ChannelID != GetChannelID(native, std.ParticipantWalletAddr, std.PartnerWalletAddr) {
		log.Error("[SetTotalDeposit] chID mismatch!")
		return utils.BYTE_FALSE, errors.NewErr("[SetTotalDeposit] chID mismatch!")
	}

	if std.SetTotalDeposit <= 0 {
		log.Error("[SetTotalDeposit] STD no larger than 0")
		return utils.BYTE_FALSE, errors.NewErr("[SetTotalDeposit] STD no larger than 0")
	}

	channel, err := GetChanInfoFromDB(native, std.ChannelID)
	if err != nil || channel == nil {
		log.Error("[SetTotalDeposit] GetChanInfoFromDB error!")
		return utils.BYTE_FALSE, errors.NewErr("[SetTotalDeposit] GetChanInfoFromDB error!")
	}

	chParticipant := WhoIsParticipant(std.ParticipantWalletAddr, channel.Participant1.WalletAddr,
		&channel.Participant1, &channel.Participant2)
	chPartner := WhoIsParticipant(std.PartnerWalletAddr, channel.Participant1.WalletAddr,
		&channel.Participant1, &channel.Participant2)

	addedDeposit := std.SetTotalDeposit - chParticipant.Deposit //TODO:Balance check
	if addedDeposit <= 0 {
		log.Error("[MPay Contract][SetTotalDeposit] Added_deposit lt 0!")
		return utils.BYTE_FALSE, errors.NewErr("[MPay Contract][SetTotalDeposit] Added_deposit lt 0!")
	}

	if channel.ChannelState != Opened {
		log.Error("[MPay Contract][SetTotalDeposit] Channel state is not openning!")
		return utils.BYTE_FALSE, errors.NewErr("[MPay Contract][SetTotalDeposit] Channel state is not openning!")
	}

	//Underflow check; we use <= because added_deposit == total_deposit for the first Deposit
	if addedDeposit > std.SetTotalDeposit {
		log.Error("[SetTotalDeposit] addedDeposit> SetTotalDeposit")
		return utils.BYTE_FALSE, errors.NewErr("[SetTotalDeposit] addedDeposit> SetTotalDeposit")
	}

	if chParticipant.Deposit+addedDeposit != std.SetTotalDeposit {
		log.Error("[SetTotalDeposit] SetTotalDeposit OverFlow")
		return utils.BYTE_FALSE, errors.NewErr("[SetTotalDeposit] SetTotalDeposit OverFlow")
	}

	chParticipant.Deposit = std.SetTotalDeposit
	//Overflow check
	chanDeposit := chParticipant.Deposit + chPartner.Deposit
	if chanDeposit < chParticipant.Deposit {
		log.Error("chanDeposit overflow check error")
		return utils.BYTE_FALSE, errors.NewErr("chanDeposit overflow check error")
	}

	err = appCallTransfer(native, utils.UsdtContractAddress, std.ParticipantWalletAddr, contract, addedDeposit)
	if err != nil {
		log.Error("[SetTotalDeposit] appCallTransfer error")
		return utils.BYTE_FALSE, errors.NewErr("[SetTotalDeposit] appCallTransfer error")
	}

	err = PutChannelInfoToDB(native, *channel)
	if err != nil {
		log.Error("[SetTotalDeposit] PutChannelInfoToDB error")
		return utils.BYTE_FALSE, errors.NewErr("[SetTotalDeposit] PutChannelInfoToDB error")
	}

	var stdEvent channelNewDepositEvent
	stdEvent.totalDeposit = std.SetTotalDeposit
	stdEvent.channelIdentifier = std.ChannelID
	stdEvent.participant = std.ParticipantWalletAddr
	ChannelNewDepositEvent(native, stdEvent, []common.Address{chParticipant.WalletAddr, chPartner.WalletAddr})

	return utils.BYTE_TRUE, nil
}

func SetTotalWithdraw(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	var withDraw WithDraw
	source := common.NewZeroCopySource(native.Input)
	err := withDraw.Deserialization(source)
	if err != nil {
		log.Error("[SetTotalWithdraw] WithDrawInfo deserialization error!")
		return utils.BYTE_FALSE, errors.NewErr("[SetTotalWithdraw] WithDrawInfo deserialization error!")
	}

	participantSignValue, err := signature.Deserialize(withDraw.ParticipantSig)
	if err != nil {
		log.Error("[SetTotalWithdraw] Participant signature deserialize error")
		return utils.BYTE_FALSE, errors.NewErr("[SetTotalWithdraw] Participant signature deserialize error")
	}

	participantPubKey, err := keypair.DeserializePublicKey(withDraw.ParticipantPubKey)
	if err != nil {
		log.Error("[SetTotalWithdraw] Participant deserialize PublicKey error")
		return utils.BYTE_FALSE, errors.NewErr("[SetTotalWithdraw] Participant deserialize PublicKey error")
	}

	msgHash := WithDrawMessageBundleHash(withDraw.ChannelID, withDraw.Participant, withDraw.TotalWithdraw)
	participantRes := signature.Verify(participantPubKey, msgHash[:], participantSignValue)
	if !participantRes {
		log.Error("[SetTotalWithdraw] Participant sig verify error!")
		return utils.BYTE_FALSE, errors.NewErr("[SetTotalWithdraw] Participant sig verify error!")
	}

	partnerSignValue, err := signature.Deserialize(withDraw.PartnerSig)
	if err != nil {
		log.Error("[MPay Contract][SetTotalWithdraw] Partner signature deserialize error")
		return utils.BYTE_FALSE, errors.NewErr("[MPay Contract][SetTotalWithdraw] Partner signature deserialize error")
	}

	partnerPubKey, err := keypair.DeserializePublicKey(withDraw.PartnerPubKey)
	if err != nil {
		log.Error("[MPay Contract][SetTotalWithdraw] Partner deserialize PublicKey error")
		return utils.BYTE_FALSE, errors.NewErr("[MPay Contract][SetTotalWithdraw] Partner deserialize PublicKey error")
	}

	partnerRes := signature.Verify(partnerPubKey, msgHash[:], partnerSignValue)
	if !partnerRes {
		log.Error("[MPay Contract][SetTotalWithdraw] Partner sig verify error!")
		return utils.BYTE_FALSE, errors.NewErr("[MPay Contract][SetTotalWithdraw] Partner sig verify error!")
	}

	var currentWithdraw uint64
	if withDraw.TotalWithdraw <= 0 {
		log.Error("[MPay Contract][SetTotalWithdraw] TotalWithdraw lq 0!")
		return utils.BYTE_FALSE, errors.NewErr("[MPay Contract][SetTotalWithdraw] TotalWithdraw lq 0!")
	}

	if withDraw.ChannelID != GetChannelID(native, withDraw.Participant, withDraw.Partner) {
		log.Error("[MPay Contract][SetTotalWithdraw] ChannelID mismatch!")
		return utils.BYTE_FALSE, errors.NewErr("[MPay Contract][SetTotalWithdraw] ChannelID mismatch!")
	}

	chanInfo, err := GetChanInfoFromDB(native, withDraw.ChannelID)
	if err != nil {
		log.Error("[MPay Contract][SetTotalWithdraw] GetChanInfoFromDB error!")
		return utils.BYTE_FALSE, errors.NewErr("[MPay Contract][SetTotalWithdraw] GetChanInfoFromDB error!")
	}

	chParticipant := WhoIsParticipant(withDraw.Participant, chanInfo.Participant1.WalletAddr,
		&chanInfo.Participant1, &chanInfo.Participant2)

	currentWithdraw = withDraw.TotalWithdraw - chParticipant.WithDrawAmount
	if currentWithdraw <= 0 {
		log.Error("[MPay Contract][SetTotalWithdraw] current_withdraw lt 0!")
		return utils.BYTE_FALSE, errors.NewErr("[MPay Contract][SetTotalWithdraw] current_withdraw lt 0!")
	}

	if chanInfo.Participant1.Nonce != 0 || chanInfo.Participant2.Nonce != 0 {
		log.Error("[MPay Contract][SetTotalWithdraw] participant Nonce nq 0")
		return utils.BYTE_FALSE, errors.NewErr("[MPay Contract][SetTotalWithdraw] participant Nonce nq 0")
	}

	//state change
	chParticipant.WithDrawAmount = withDraw.TotalWithdraw //不用更新channel deposit和participant.Deposit
	err = appCallTransfer(native, utils.UsdtContractAddress, contract, withDraw.Participant, currentWithdraw)
	if err != nil {
		log.Error("[SetTotalWithdraw] Call transfer ont error!")
		return utils.BYTE_FALSE, errors.NewErr("[SetTotalWithdraw] Call transfer ont error!")
	}

	err = PutChannelInfoToDB(native, *chanInfo)
	if err != nil {
		log.Error("[SetTotalWithdraw] PutChanInfoFromDB error!")
		return utils.BYTE_FALSE, errors.NewErr("[SetTotalWithdraw] PutChanInfoFromDB error!")
	}

	var withdrawEvent channelWithdrawEvent
	withdrawEvent.totalWithdraw = withDraw.TotalWithdraw
	withdrawEvent.channelIdentifier = withDraw.ChannelID
	withdrawEvent.participant = withDraw.Participant
	ChannelWithdrawEvent(native, withdrawEvent, []common.Address{chanInfo.Participant1.WalletAddr, chanInfo.Participant2.WalletAddr})

	return utils.BYTE_TRUE, nil
}

func CloseChannel(native *native.NativeService) ([]byte, error) {
	var closeChannelInfo CloseChannelInfo
	source := common.NewZeroCopySource(native.Input)
	err := closeChannelInfo.Deserialization(source)
	if err != nil {
		log.Error("[CloseChannel] closeChannelInfo deserialization error!")
		return utils.BYTE_FALSE, errors.NewErr("[CloseChannel] closeChannelInfo deserialization error!")
	}

	channelID := GetChannelID(native, closeChannelInfo.ParticipantAddress, closeChannelInfo.PartnerAddress)
	if closeChannelInfo.ChannelID != channelID {
		log.Error("CloseChannel CHID mismatch!")
		return utils.BYTE_FALSE, errors.NewErr("CloseChannel CHID mismatch!")
	}

	chanInfo, err := GetChanInfoFromDB(native, channelID)
	if err != nil || chanInfo == nil {
		log.Error("[CloseChannel] GetChanInfoFromDB error!")
		return utils.BYTE_FALSE, errors.NewErr("[CloseChannel] GetChanInfoFromDB error!")
	}

	if chanInfo.ChannelState != Opened {
		log.Error("CloseChannel ChannelState is not opened!")
		return utils.BYTE_FALSE, errors.NewErr("CloseChannel ChannelState is not opened!")
	}

	chParticipant := WhoIsParticipant(closeChannelInfo.ParticipantAddress, chanInfo.Participant1.WalletAddr, &chanInfo.Participant1, &chanInfo.Participant2)
	chanInfo.ChannelState = Closed
	chParticipant.IsCloser = true
	chanInfo.SettleBlockHeight += uint64(native.Height)
	if closeChannelInfo.Nonce > 0 {
		partnerPubKey, err := keypair.DeserializePublicKey(closeChannelInfo.PartnerPubKey)
		if err != nil {
			log.Error("[closeChannel] Partner deserialize PublicKey error")
			return utils.BYTE_FALSE, errors.NewErr("[closeChannel] Partner deserialize PublicKey error")
		}

		partnerSignValue, err := signature.Deserialize(closeChannelInfo.PartnerSignature)
		if err != nil {
			log.Error("[closeChannel] Participant2 signature deserialize error")
			return utils.BYTE_FALSE, errors.NewErr("[closeChannel] Participant2 signature deserialize error")
		}

		msgHash := ClosedMessageBundleHash(channelID, closeChannelInfo.BalanceHash, closeChannelInfo.Nonce, closeChannelInfo.AdditionalHash)

		Res := signature.Verify(partnerPubKey, msgHash[:], partnerSignValue)
		if !Res {
			log.Error("[closeChannel] Partner sig verify error!")
			return utils.BYTE_FALSE, errors.NewErr("[closeChannel] Partner sig verify error!")
		}

		err = UpdateBalanceProofData(chanInfo, closeChannelInfo.PartnerAddress, closeChannelInfo.Nonce, closeChannelInfo.BalanceHash)
		if err != nil {
			log.Error("[closeChannel] UpdateBalanceProofData error!")
			return utils.BYTE_FALSE, errors.NewErr("[closeChannel] UpdateBalanceProofData error!")
		}
	}

	if err = PutChannelInfoToDB(native, *chanInfo); err != nil {
		log.Error("[CloseChannel] PutChannelInfoToDB error!")
		return utils.BYTE_FALSE, errors.NewErr("[CloseChannel] PutChannelInfoToDB error!")
	}

	var closeChanEvent channelCloseEvent
	closeChanEvent.nonce = closeChannelInfo.Nonce
	closeChanEvent.closingParticipant = closeChannelInfo.ParticipantAddress
	closeChanEvent.channelID = closeChannelInfo.ChannelID
	ChannelCloseEvent(native, closeChanEvent, []common.Address{chanInfo.Participant1.WalletAddr, chanInfo.Participant2.WalletAddr})

	return utils.BYTE_TRUE, nil
}

func CooperativeSettle(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	var coSettle CooperativeSettleInfo
	source := common.NewZeroCopySource(native.Input)
	err := coSettle.Deserialization(source)
	if err != nil {
		log.Error("[CooperativeSettle] CooperativeSettleInfo deserialization error!")
		return utils.BYTE_FALSE, errors.NewErr("[CooperativeSettle] CooperativeSettleInfo deserialization error!")
	}

	participant1PubKey, err := keypair.DeserializePublicKey(coSettle.Participant1PubKey)
	if err != nil || participant1PubKey == nil {
		log.Error("[CooperativeSettle] Participant1 deserialize PublicKey error")
		return utils.BYTE_FALSE, errors.NewErr("[CooperativeSettle] Participant1 deserialize PublicKey error")
	}

	participant1SignValue, err := signature.Deserialize(coSettle.Participant1Signature)
	if err != nil || participant1SignValue == nil {
		log.Error("[CooperativeSettle] Participant1 signature deserialize error")
		return utils.BYTE_FALSE, errors.NewErr("[CooperativeSettle] Participant1 signature deserialize error")
	}

	participant2PubKey, err := keypair.DeserializePublicKey(coSettle.Participant2PubKey)
	if err != nil || participant2PubKey == nil {
		log.Error("[CooperativeSettle] Participant2 deserialize PublicKey error")
		return utils.BYTE_FALSE, errors.NewErr("[CooperativeSettle] Participant2 deserialize PublicKey error")
	}

	participant2SignValue, err := signature.Deserialize(coSettle.Participant2Signature)
	if err != nil || participant2SignValue == nil {
		log.Error("[CooperativeSettle] Participant2 signature deserialize error")
		return utils.BYTE_FALSE, errors.NewErr("[CooperativeSettle] Participant2 signature deserialize error")
	}

	channelID := GetChannelID(native, coSettle.Participant1Address, coSettle.Participant2Address)
	if coSettle.ChannelID != channelID {
		log.Error("CooperativeSettle CHID mismatch!")
		return utils.BYTE_FALSE, errors.NewErr("CooperativeSettle CHID mismatch!")
	}

	pairHash := GetParticipantHash(coSettle.Participant1Address, coSettle.Participant2Address)

	chanInfo, err := GetChanInfoFromDB(native, channelID)
	if err != nil || chanInfo == nil {
		log.Error("[CooperativeSettle] GetChanInfoFromDB error!")
		return utils.BYTE_FALSE, errors.NewErr("[CooperativeSettle] GetChanInfoFromDB error!")
	}

	if chanInfo.ChannelState != Opened {
		log.Error("CooperativeSettle ChannelState is not opened!")
		return utils.BYTE_FALSE, errors.NewErr("CooperativeSettle ChannelState is not opened!")
	}

	msgHash := CoSettleMessageBundleHash(channelID, coSettle.Participant1Address,
		coSettle.Participant1Balance, coSettle.Participant2Address, coSettle.Participant2Balance)
	participant1Res := signature.Verify(participant1PubKey, msgHash[:], participant1SignValue)
	if !participant1Res {
		log.Error("[CooperativeSettle] Participant1 sig verify error!")
		return utils.BYTE_FALSE, errors.NewErr("[CooperativeSettle] Participant1 sig verify error!")
	}

	participant2Res := signature.Verify(participant2PubKey, msgHash[:], participant2SignValue)
	if !participant2Res {
		log.Error("[CooperativeSettle] Participant2 sig verify error!")
		return utils.BYTE_FALSE, errors.NewErr("[CooperativeSettle] Participant2 sig verify error!")
	}

	totalAvailableDeposit := GetChannelAvailableDeposit(chanInfo.Participant1, chanInfo.Participant2)
	if totalAvailableDeposit != (coSettle.Participant1Balance + coSettle.Participant2Balance) {
		log.Error("CooperativeSettle TAD!=SUM(P.Balance)")
		return utils.BYTE_FALSE, errors.NewErr("CooperativeSettle TAD!=SUM(P.Balance)")
	}

	if coSettle.Participant1Balance > (coSettle.Participant1Balance + coSettle.Participant2Balance) {
		log.Error("P1.Balance OverFlow Check fail!")
		return utils.BYTE_FALSE, errors.NewErr("P1.Balance OverFlow Check fail!")
	}

	participant1Addr := chanInfo.Participant1.WalletAddr
	participant2Addr := chanInfo.Participant2.WalletAddr
	//Del channel data from DB before transfer
	DeleteChanInfoFromDB(native, channelID)
	utils.DelStorageItem(native, pairHash[:])

	if coSettle.Participant1Balance > 0 {
		err = appCallTransfer(native, utils.UsdtContractAddress, contract,
			coSettle.Participant1Address, coSettle.Participant1Balance)
		if err != nil {
			log.Error("[CooperativeSettle] appCallTransfer to participant1 error!")
			return utils.BYTE_FALSE, errors.NewErr("[CooperativeSettle] appCallTransfer to participant1 error!")
		}
	}

	if coSettle.Participant2Balance > 0 {
		err = appCallTransfer(native, utils.UsdtContractAddress, contract,
			coSettle.Participant2Address, coSettle.Participant2Balance)
		if err != nil {
			log.Error("[CooperativeSettle] appCallTransfer to participant2 error!")
			return utils.BYTE_FALSE, errors.NewErr("[CooperativeSettle] appCallTransfer to participant2 error!")
		}
	}

	var chanSettle channelSettledEvent
	chanSettle.channelID = channelID
	chanSettle.participant1_amount = coSettle.Participant1Balance
	chanSettle.participant2_amount = coSettle.Participant2Balance
	ChannelCooperativeSettledEvent(native, chanSettle, []common.Address{participant1Addr, participant2Addr})

	return utils.BYTE_TRUE, nil
}

func UpdateNonClosingBalanceProof(native *native.NativeService) ([]byte, error) {
	var updateNonCloseBPF UpdateNonCloseBalanceProof

	source := common.NewZeroCopySource(native.Input)
	err := updateNonCloseBPF.Deserialization(source)
	if err != nil {
		log.Error("[UpdateNonClosingBalanceProof] UpdateNonCloseBalanceProof deserialization error!")
		return utils.BYTE_FALSE, errors.NewErr("[UpdateNonClosingBalanceProof] UpdateNonCloseBalanceProof deserialization error!")
	}

	if updateNonCloseBPF.ChanID != GetChannelID(native, updateNonCloseBPF.CloseParticipant, updateNonCloseBPF.NonCloseParticipant) {
		log.Error("[UpdateNonClosingBalanceProof] ChanID mismatch!")
		return utils.BYTE_FALSE, errors.NewErr("[UpdateNonClosingBalanceProof] ChanID mismatch!")
	}

	if updateNonCloseBPF.Nonce <= 0 {
		log.Error("[UpdateNonClosingBalanceProof]Nonce lt zero!")
		return utils.BYTE_FALSE, errors.NewErr("[UpdateNonClosingBalanceProof]Nonce lt zero!")
	}

	chInfo, err := GetChanInfoFromDB(native, updateNonCloseBPF.ChanID)
	if err != nil || chInfo == nil {
		log.Error("[UpdateNonClosingBalanceProof] GetChanInfoFromDB error")
		return utils.BYTE_FALSE, errors.NewErr("[UpdateNonClosingBalanceProof] GetChanInfoFromDB error")
	}

	if chInfo.ChannelState != Closed {
		log.Error("[UpdateNonClosingBalanceProof]ChannelState neq Closed!")
		return utils.BYTE_FALSE, errors.NewErr("[UpdateNonClosingBalanceProof]ChannelState neq Closed!")
	}

	if chInfo.SettleBlockHeight < uint64(native.Height) {
		log.Error("[UpdateNonClosingBalanceProof] SettleBlockHeight lt current BlockHeight!")
		return utils.BYTE_FALSE, errors.NewErr("[UpdateNonClosingBalanceProof] SettleBlockHeight lt current BlockHeight!")
	}

	closePubKey, err := keypair.DeserializePublicKey(updateNonCloseBPF.ClosePubKey)
	if err != nil || closePubKey == nil {
		log.Error("[UpdateNonClosingBalanceProof] closePubKey deserialize error")
		return utils.BYTE_FALSE, errors.NewErr("[UpdateNonClosingBalanceProof] closePubKey deserialize error")
	}

	closeSig, err := signature.Deserialize(updateNonCloseBPF.CloseSignature)
	if err != nil || closeSig == nil {
		log.Error("[UpdateNonClosingBalanceProof] closeSig deserialize error")
		return utils.BYTE_FALSE, errors.NewErr("[UpdateNonClosingBalanceProof] closeSig deserialize error")
	}

	nonClosePubKey, err := keypair.DeserializePublicKey(updateNonCloseBPF.NonClosePubKey)
	if err != nil || nonClosePubKey == nil {
		log.Error("[UpdateNonClosingBalanceProof] nonClosePubKey deserialize error")
		return utils.BYTE_FALSE, errors.NewErr("[UpdateNonClosingBalanceProof] nonClosePubKey deserialize error")
	}

	nonCloseSig, err := signature.Deserialize(updateNonCloseBPF.NonCloseSignature)
	if err != nil || nonCloseSig == nil {
		log.Error("[UpdateNonClosingBalanceProof] nonCloseSig deserialize error")
		return utils.BYTE_FALSE, errors.NewErr("[UpdateNonClosingBalanceProof] nonCloseSig deserialize error")
	}

	//NonClose msgHash
	nonCloseMsgHash := BalanceProofUpdateMessageBundleHash(
		updateNonCloseBPF.ChanID,
		updateNonCloseBPF.BalanceHash,
		updateNonCloseBPF.Nonce,
		updateNonCloseBPF.AdditionalHash,
		updateNonCloseBPF.CloseSignature,
	)
	//close msgHash
	closeMsgHash := ClosedMessageBundleHash(
		updateNonCloseBPF.ChanID,
		updateNonCloseBPF.BalanceHash,
		updateNonCloseBPF.Nonce,
		updateNonCloseBPF.AdditionalHash,
	)
	//verify NonClose sig
	nonCloseRes := signature.Verify(nonClosePubKey, nonCloseMsgHash[:], nonCloseSig)
	if !nonCloseRes {
		log.Error("[UpdateNonClosingBalanceProof] nonCloseSig verify error!")
		return utils.BYTE_FALSE, errors.NewErr("[UpdateNonClosingBalanceProof] nonCloseSig verify error!")
	}

	//verify Close sig
	closeRes := signature.Verify(closePubKey, closeMsgHash[:], closeSig)
	if !closeRes {
		log.Error("[UpdateNonClosingBalanceProof] CloseSig verify error!")
		return utils.BYTE_FALSE, errors.NewErr("[UpdateNonClosingBalanceProof] CloseSig verify error!")
	}

	closeParticipant := WhoIsParticipant(updateNonCloseBPF.CloseParticipant,
		chInfo.Participant1.WalletAddr, &chInfo.Participant1, &chInfo.Participant2)
	if !closeParticipant.IsCloser {
		log.Error("closeParticipant is not closer!")
		return utils.BYTE_FALSE, errors.NewErr("closeParticipant is not closer!")
	}

	err = UpdateBalanceProofData(chInfo, updateNonCloseBPF.CloseParticipant, updateNonCloseBPF.Nonce, updateNonCloseBPF.BalanceHash)
	if err != nil {
		log.Error("[UpdateNonClosingBalanceProof] UpdateBalanceProofData error!")
		return utils.BYTE_FALSE, errors.NewErr("[UpdateNonClosingBalanceProof] UpdateBalanceProofData error!")
	}

	err = PutChannelInfoToDB(native, *chInfo)
	if err != nil {
		log.Error("[UpdateNonClosingBalanceProof] PutChanInfoFromDB error!")
		return utils.BYTE_FALSE, errors.NewErr("[UpdateNonClosingBalanceProof] PutChanInfoFromDB error!")
	}

	NonClosingBPFUpdateEvent(native, updateNonCloseBPF.ChanID, updateNonCloseBPF.CloseParticipant,
		updateNonCloseBPF.Nonce, []common.Address{chInfo.Participant1.WalletAddr, chInfo.Participant2.WalletAddr})
	return utils.BYTE_TRUE, nil
}

func SettleChannel(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	var settleChInfo SettleChannelInfo
	source := common.NewZeroCopySource(native.Input)
	if err := settleChInfo.Deserialization(source); err != nil {
		log.Error("[SettleChannel] settleChInfo deserialize error!")
		return utils.BYTE_FALSE, errors.NewErr("[SettleChannel] settleChInfo deserialize error!")
	}

	chID := GetChannelID(native, settleChInfo.Participant1, settleChInfo.Participant2)
	if settleChInfo.ChanID != chID {
		log.Error("[MPSettleChannel][ChanID] mismatch!")
		return utils.BYTE_FALSE, errors.NewErr("[MPSettleChannel][ChanID] mismatch!")
	}

	pairHash := GetParticipantHash(settleChInfo.Participant1, settleChInfo.Participant2)

	chanInfo, err := GetChanInfoFromDB(native, chID)
	if err != nil || chanInfo == nil {
		log.Error("[SettleChannel] GetChanInfoFromDB error!")
		return utils.BYTE_FALSE, errors.NewErr("[SettleChannel] GetChanInfoFromDB error!")
	}

	if chanInfo.ChannelState != Closed {
		log.Error("[MPSettleChannel] ChannelState is not closed!")
		return utils.BYTE_FALSE, errors.NewErr("[MPSettleChannel] ChannelState is not closed!")
	}

	if chanInfo.SettleBlockHeight >= uint64(native.Height) {
		log.Error("[MPSettleChannel]SettleBlockHeight gt current blockHeight!")
		return utils.BYTE_FALSE, errors.NewErr("[MPSettleChannel]SettleBlockHeight gt current blockHeight!")
	}

	chP1 := WhoIsParticipant(settleChInfo.Participant1, chanInfo.Participant1.WalletAddr, &chanInfo.Participant1, &chanInfo.Participant2)
	chP2 := WhoIsParticipant(settleChInfo.Participant2, chanInfo.Participant1.WalletAddr, &chanInfo.Participant1, &chanInfo.Participant2)

	if !verifyBalanceHashData(*chP1, settleChInfo.P1TransferredAmount, settleChInfo.P1LockedAmount, settleChInfo.P1LocksRoot) {
		log.Error("[MPSettleChannel] verifyBalanceHashData of P1 Failed!")
		return utils.BYTE_FALSE, errors.NewErr("[MPSettleChannel] verifyBalanceHashData of P1 Failed!")
	}

	if !verifyBalanceHashData(*chP2, settleChInfo.P2TransferredAmount, settleChInfo.P2LockedAmount, settleChInfo.P2LocksRoot) {
		log.Error("[MPSettleChannel] verifyBalanceHashData of P2 Failed!")
		return utils.BYTE_FALSE, errors.NewErr("[MPSettleChannel] verifyBalanceHashData of P2 Failed!")
	}

	settleChInfo.P1TransferredAmount, settleChInfo.P2TransferredAmount,
		settleChInfo.P1LockedAmount, settleChInfo.P2LockedAmount, err =
		getSettleTransferAmounts(*chP1, settleChInfo.P1TransferredAmount, settleChInfo.P1LockedAmount,
			*chP2, settleChInfo.P2TransferredAmount, settleChInfo.P2LockedAmount)
	if err != nil {
		log.Error("[MPSettleChannel] getSettleTransferAmounts error!")
		return utils.BYTE_FALSE, errors.NewErr("[MPSettleChannel] getSettleTransferAmounts error!")
	}

	participant1Addr := chanInfo.Participant1.WalletAddr
	participant2Addr := chanInfo.Participant2.WalletAddr

	//Del channel data from DB before transfer
	DeleteChanInfoFromDB(native, chID)
	utils.DelStorageItem(native, pairHash[:])

	err = storeUnlockData(native, settleChInfo.ChanID, settleChInfo.Participant1,
		settleChInfo.Participant2, settleChInfo.P1LockedAmount, settleChInfo.P1LocksRoot)
	if err != nil {
		log.Error("[MPSettleChannel] store unlock data for participant1 error!")
		return utils.BYTE_FALSE, errors.NewErr("[MPSettleChannel] store unlock data for participant1 error!")
	}

	err = storeUnlockData(native, settleChInfo.ChanID, settleChInfo.Participant2,
		settleChInfo.Participant1, settleChInfo.P2LockedAmount, settleChInfo.P2LocksRoot)
	if err != nil {
		log.Error("[MPSettleChannel] store unlock data for participant2 error!")
		return utils.BYTE_FALSE, errors.NewErr("[MPSettleChannel] store unlock data for participant2 error!")
	}

	if settleChInfo.P1TransferredAmount > 0 {
		err = appCallTransfer(native, utils.UsdtContractAddress, contract, settleChInfo.Participant1, settleChInfo.P1TransferredAmount)
		if err != nil {
			log.Error("[MPSettleChannel] appCallTransfer to participant1 error!")
			return utils.BYTE_FALSE, errors.NewErr("[MPSettleChannel] appCallTransfer to participant1 error!")
		}
	}
	if settleChInfo.P2TransferredAmount > 0 {
		err = appCallTransfer(native, utils.UsdtContractAddress, contract, settleChInfo.Participant2, settleChInfo.P2TransferredAmount)
		if err != nil {
			log.Error("[MPSettleChannel] appCallTransfer to participant2 error!")
			return utils.BYTE_FALSE, errors.NewErr("[MPSettleChannel] appCallTransfer to participant2 error!")
		}
	}

	var chanSettledEvent channelSettledEvent
	chanSettledEvent.channelID = chID
	chanSettledEvent.participant1_amount = settleChInfo.P1TransferredAmount
	chanSettledEvent.participant2_amount = settleChInfo.P2TransferredAmount
	ChannelSettledEvent(native, chanSettledEvent, []common.Address{participant1Addr, participant2Addr})

	return utils.BYTE_TRUE, nil
}

//Todo: to improvement
func GetChannelInfo(native *native.NativeService) ([]byte, error) {
	var getChInfoParam GetChanInfo
	var channel *ChannelInfo

	source := common.NewZeroCopySource(native.Input)
	getChInfoParam.Deserialization(source)
	channel, err := GetChanInfoFromDB(native, getChInfoParam.ChannelID)
	if err != nil || channel == nil {
		channel = new(ChannelInfo)
		channel.SettleBlockHeight = 0
		channel.ChannelState = Settled
	}

	sink := common.NewZeroCopySink(nil)
	utils.EncodeVarUint(sink, channel.ChannelID)
	utils.EncodeVarUint(sink, channel.SettleBlockHeight)
	utils.EncodeVarUint(sink, channel.ChannelState)
	utils.EncodeAddress(sink, channel.Participant1.WalletAddr)
	utils.EncodeBool(sink, channel.Participant1.IsCloser)
	utils.EncodeAddress(sink, channel.Participant2.WalletAddr)
	utils.EncodeBool(sink, channel.Participant2.IsCloser)
	return sink.Bytes(), nil
}

func GetAllOpenChannels(native *native.NativeService) ([]byte, error) {
	var channelCounter uint64

	channelCounterItem, err := utils.GetStorageItem(native, []byte("channelCounter"))
	if err != nil || channelCounterItem == nil {
		channelCounter = 100
		buff := new(bytes.Buffer)
		utils.WriteVarUint(buff, channelCounter)
		utils.PutBytes(native, []byte("channelCounter"), buff.Bytes())
	}

	if channelCounterItem != nil {
		reader := bytes.NewReader(channelCounterItem.Value)
		channelCounter, err = utils.ReadVarUint(reader)
		if err != nil {
			log.Error("[GetChannelCounter] channelCounter ReadVarUint error")
			return utils.BYTE_FALSE, errors.NewErr("[GetChannelCounter] channelCounter ReadVarUint error")
		}
	}

	var openChannels AllChannels
	for chanId := uint64(101); chanId < channelCounter+1; chanId++ {
		channelInfo, err := GetChanInfoFromDB(native, chanId)
		if err == nil && channelInfo != nil {
			if channelInfo.ChannelState == Opened {
				participant := Participants{
					ChannelID: channelInfo.ChannelID,
					Part1Addr: channelInfo.Participant1.WalletAddr,
					Part2Addr: channelInfo.Participant2.WalletAddr,
				}
				openChannels.Participants = append(openChannels.Participants, participant)
				openChannels.ParticipantNum++
			}
		}
	}
	buf := new(bytes.Buffer)
	err = openChannels.Serialize(buf)
	if err != nil {
		log.Error(err.Error())
		return utils.BYTE_FALSE, errors.NewErr("[GetChannelCounter] openChannels serialization error")
	}
	return buf.Bytes(), nil
}

//Todo: to add unlock_data
func GetChannelParticipantInfo(native *native.NativeService) ([]byte, error) {
	var getCPInfoParam GetChanInfo

	participantInfo := new(Participant)
	source := common.NewZeroCopySource(native.Input)
	getCPInfoParam.Deserialization(source)
	chanInfo, err := GetChanInfoFromDB(native, getCPInfoParam.ChannelID)
	if err == nil && chanInfo != nil {
		participantInfo = WhoIsParticipant(getCPInfoParam.Participant1, chanInfo.Participant1.WalletAddr,
			&chanInfo.Participant1, &chanInfo.Participant2)
	} else {
		participantInfo.WalletAddr = getCPInfoParam.Participant1
	}

	unlockKey, err := getUnlockId(getCPInfoParam.ChannelID, getCPInfoParam.Participant1, getCPInfoParam.Participant2)
	if err != nil {
		log.Error("[GetChannelParticipantInfo] get unlock id error")
		return utils.BYTE_FALSE, errors.NewErr("[GetChannelParticipantInfo] get unlock id error")
	}

	unlockItem, err := utils.GetStorageItem(native, unlockKey)
	log.Debugf("GetStorageItem for channel : %v, unlockKey : %v, unlockItem : %v, err : %v\n", getCPInfoParam.ChannelID, unlockKey, unlockItem, err)
	if err == nil && unlockItem != nil {
		unlockData := new(UnlockDataInfo)
		unlockDataInfoSource := common.NewZeroCopySource(unlockItem.Value)
		err = unlockData.Deserialization(unlockDataInfoSource)
		if err != nil {
			log.Error("[GetChannelParticipantInfo] unlock data info deserialize error!")
			return utils.BYTE_FALSE, errors.NewErr("[GetChannelParticipantInfo] unlock data info deserialize error!")
		}

		participantInfo.LocksRoot = unlockData.LocksRoot
		participantInfo.LockedAmount = unlockData.LockedAmount
		log.Debugf("locksRoot : %v, lockedAmount : %d\n", participantInfo.LocksRoot, participantInfo.LockedAmount)
	}

	sink := common.NewZeroCopySink(nil)
	participantInfo.Serialization(sink)
	return sink.Bytes(), nil
}

func GetChannelIdentifier(native *native.NativeService) ([]byte, error) {
	var getChID GetChannelId
	source := common.NewZeroCopySource(native.Input)
	getChID.Deserialization(source)
	chid := GetChannelID(native, getChID.Participant1WalletAddr, getChID.Participant2WalletAddr)
	return util.Int64ToBytes(chid), nil
}

func FastTransfer(native *native.NativeService) ([]byte, error) {

	var info TransferInfo
	source := common.NewZeroCopySource(native.Input)
	err := info.Deserialization(source)
	if err != nil {
		log.Error("[FastTransfer] TransferInfo deserialization error!")
		return utils.BYTE_FALSE, errors.NewErr("[FastTransfer] ChannelInfo deserialization error!")
	}

	if !native.ContextRef.CheckWitness(info.From) {
		return utils.BYTE_FALSE, errors.NewErr("[FastTransfer] CheckWitness error!")
	}

	if info.Amount == 0 {
		return utils.BYTE_TRUE, nil
	}

	err = appCallTransfer(native, utils.UsdtContractAddress, info.From, info.To, info.Amount)
	if err != nil {
		log.Errorf("[FastTransfer] appCallTransfer error %s", err)
		return utils.BYTE_FALSE, errors.NewErr("[FastTransfer] appCallTransfer error")
	}

	var evt fastTransferEvent
	evt.paymentId = info.PaymentId
	evt.from = info.From
	evt.to = info.To
	evt.amount = info.Amount
	evt.asset = utils.UsdtContractAddress
	NewFastTransferEvent(native, evt, []common.Address{info.From, info.To})

	return utils.BYTE_TRUE, nil
}

func appCallTransfer(native *native.NativeService, contract common.Address, from common.Address, to common.Address, amount uint64) error {
	var sts []usdt.State
	sts = append(sts, usdt.State{
		From:  from,
		To:    to,
		Value: amount,
	})
	transfers := usdt.Transfers{
		States: sts,
	}
	sink := common.NewZeroCopySink(nil)
	transfers.Serialization(sink)

	if _, err := native.NativeCall(contract, "transfer", sink.Bytes()); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "appCallTransfer, appCall error!")
	}
	return nil
}
