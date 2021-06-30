package film

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/dns"
	"github.com/saveio/themis/smartcontract/service/native/savefs"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

func FilmPublish(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	buf, err := utils.DecodeBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] FilmPublish deserialize error!")
	}
	params := make([]interface{}, 0)
	err = json.Unmarshal(buf, &params)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] FilmPublish unmarshal error!")
	}
	log.Debugf("len(params) %d\n", len(params))
	if len(params) < 10 {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] FilmPublish unmarshal error!")
	}
	filmInfo := &FilmInfo{
		Cover:       []byte(getStringValue(params[0])),
		Url:         []byte(getStringValue(params[1])),
		Name:        []byte(getStringValue(params[2])),
		Desc:        []byte(getStringValue(params[3])),
		Type:        getUint64Value(params[4]),
		ReleaseYear: getUint64Value(params[5]),
		Language:    []byte(getStringValue(params[6])),
		Region:      []byte(getStringValue(params[7])),
		Price:       getUint64Value(params[8]),
		Available:   getBoolValue(params[9]),
	}
	log.Debugf("params[8] %v, %T, %v %v\n", params[8], params[8], getStringValue(params[8]), getUint64Value(getStringValue(params[8])))
	log.Debugf("filmInfo.cover: %v\n", filmInfo.Cover)
	log.Debugf("filmInfo.url: %s\n", filmInfo.Url)
	log.Debugf("filmInfo.Name: %v\n", filmInfo.Name)
	log.Debugf("filmInfo.Desc: %v\n", filmInfo.Desc)
	log.Debugf("filmInfo.Available: %v\n", filmInfo.Available)
	log.Debugf("filmInfo.Type: %v\n", filmInfo.Type)
	log.Debugf("filmInfo.ReleaseYear: %v\n", filmInfo.ReleaseYear)
	log.Debugf("filmInfo.Language: %v\n", filmInfo.Language)
	log.Debugf("filmInfo.Region: %v\n", filmInfo.Region)
	log.Debugf("filmInfo.Price: %v\n", filmInfo.Price)
	log.Debugf("filmInfo.Available: %v\n", filmInfo.Available)

	fileHash, err := getFileHashFromUrl(native, string(filmInfo.Url))
	log.Debugf("fileHash %s\n", fileHash)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] get filehash error!" + err.Error())
	}
	filmInfo.Id = sha256.New().Sum(fileHash)
	filmInfo.Hash = fileHash
	// filmInfo.Available = true

	txHash := native.Tx.Hash()
	log.Debugf("filmId %v, hash %x\n", filmInfo.Id, common.ToArrayReverse(txHash[:]))
	fileInfo, err := getFileInfo(native, filmInfo.Hash)
	log.Debugf("fileInfo %v err %v\n", fileInfo, err)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] get fileInfo error!" + err.Error())
	}

	if !native.ContextRef.CheckWitness(fileInfo.FileOwner) {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] CheckWitness failed!")
	}
	filmInfo.Owner = fileInfo.FileOwner
	oldFilm, err := getFilmInfo(native, fileInfo.FileOwner, fileInfo.FileHash)
	if oldFilm != nil && err == nil {
		return utils.BYTE_TRUE, nil
	}
	filmInfo.FileSize = fileInfo.FileBlockNum * fileInfo.FileBlockSize
	filmInfo.RealFileSize = fileInfo.RealFileSize
	// add to user film list
	err = addFilmToUserFilmList(native, fileInfo.FileOwner, filmInfo.Hash)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] put film to user film list err!")
	}
	// add to all film list
	err = addFilmToAllFilmList(native, filmInfo, fileInfo.FileOwner)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] put film to all film list err!")
	}

	// add to film info storage
	contract := native.ContextRef.CurrentContext().ContractAddress
	filmInfoKey := GenFilmInfoKey(contract, fileInfo.FileOwner, filmInfo.Hash)
	log.Debugf("filmInfoKey: %v\n", filmInfoKey)
	bf := new(bytes.Buffer)
	if err = filmInfo.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsFileProve filminfo serialize error!")
	}
	utils.PutBytes(native, filmInfoKey, bf.Bytes())
	return utils.BYTE_TRUE, nil
}

func FilmUpdate(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	buf, err := utils.DecodeBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] FilmPublish deserialize error!")
	}
	params := make([]interface{}, 0)
	err = json.Unmarshal(buf, &params)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] FilmPublish unmarshal error!")
	}
	if len(params) < 10 {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] FilmPublish unmarshal error!")
	}
	fileHash, err := getFileHashFromUrl(native, string([]byte(getStringValue(params[1]))))
	log.Debugf("FilmUpdate %s\n", fileHash)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] get fileInfo error!" + err.Error())
	}
	fileInfo, err := getFileInfo(native, fileHash)
	log.Debugf("FilmUpdate %v err %v\n", fileInfo, err)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] get fileInfo error!" + err.Error())
	}
	item, err := getFilmInfo(native, fileInfo.FileOwner, fileInfo.FileHash)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] get fileInfo error!" + err.Error())
	}
	filmInfo := &FilmInfo{}
	r := bytes.NewReader(item)
	err = filmInfo.Deserialize(r)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] Film Deserialize err!")
	}
	filmInfo.Id = sha256.New().Sum(fileHash)
	filmInfo.Cover = []byte(getStringValue(params[0]))
	filmInfo.Url = []byte(getStringValue(params[1]))
	filmInfo.Name = []byte(getStringValue(params[2]))
	filmInfo.Desc = []byte(getStringValue(params[3]))
	filmInfo.Type = getUint64Value(params[4])
	filmInfo.ReleaseYear = getUint64Value(params[5])
	filmInfo.Language = []byte(getStringValue(params[6]))
	filmInfo.Region = []byte(getStringValue(params[7]))
	filmInfo.Price = getUint64Value(params[8])
	filmInfo.Available = getBoolValue(params[9])

	log.Debugf("filmInfo.cover: %v\n", filmInfo.Cover)
	log.Debugf("filmInfo.url: %s\n", filmInfo.Url)
	log.Debugf("filmInfo.Name: %v\n", filmInfo.Name)
	log.Debugf("filmInfo.Desc: %v\n", filmInfo.Desc)
	log.Debugf("filmInfo.Available: %v\n", filmInfo.Available)
	log.Debugf("filmInfo.Type: %v\n", filmInfo.Type)
	log.Debugf("filmInfo.ReleaseYear: %v\n", filmInfo.ReleaseYear)
	log.Debugf("filmInfo.Language: %v\n", filmInfo.Language)
	log.Debugf("filmInfo.Region: %v\n", filmInfo.Region)
	log.Debugf("filmInfo.Price: %v\n", filmInfo.Price)
	log.Debugf("filmInfo.Available: %v %v %T\n", filmInfo.Available, params[9], params[9])

	if !native.ContextRef.CheckWitness(fileInfo.FileOwner) {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] CheckWitness failed!")
	}
	filmInfo.Owner = fileInfo.FileOwner
	filmInfo.FileSize = fileInfo.FileBlockNum * fileInfo.FileBlockSize
	filmInfo.RealFileSize = fileInfo.RealFileSize

	// add to film info storage
	contract := native.ContextRef.CurrentContext().ContractAddress
	filmInfoKey := GenFilmInfoKey(contract, fileInfo.FileOwner, filmInfo.Hash)
	log.Debugf("filmInfoKey: %v\n", filmInfoKey)
	bf := new(bytes.Buffer)
	if err = filmInfo.Serialize(bf); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsFileProve filminfo serialize error!")
	}
	utils.PutBytes(native, filmInfoKey, bf.Bytes())
	return utils.BYTE_TRUE, nil
}

func BuyFilm(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	buf, err := utils.DecodeBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] FilmPublish deserialize error!")
	}
	params := make([]interface{}, 0)
	err = json.Unmarshal(buf, &params)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] FilmPublish unmarshal error!")
	}
	if len(params) < 3 {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] FilmPublish params miss error!")
	}
	fileHash := []byte(getStringValue(params[0]))
	owner, err := common.AddressFromBase58(getStringValue(params[1]))
	if err != nil {
		fmt.Println("err124", err)
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] GetAllFilmList owner AddressFromBase58 error!")
	}
	user, err := common.AddressFromBase58(getStringValue(params[2]))
	if err != nil {
		fmt.Println("err128", err)
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] GetAllFilmList user AddressFromBase58 error!")
	}
	if !native.ContextRef.CheckWitness(user) {
		fmt.Println("err133", err)
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] CheckWitness failed!")
	}

	item, err := getFilmInfo(native, owner, fileHash)
	if err != nil {
		fmt.Println("err139", err)
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] Film not exists!")
	}

	filmInfo := &FilmInfo{}
	r := bytes.NewReader(item)
	err = filmInfo.Deserialize(r)
	if err != nil {
		fmt.Println("err147", err)
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] Film Deserialize err!")
	}

	filmInfo.PaidCount++

	contract := native.ContextRef.CurrentContext().ContractAddress
	userFilmKeys := GenFilmInfoKey(contract, owner, fileHash)
	bf := new(bytes.Buffer)
	if err = filmInfo.Serialize(bf); err != nil {
		fmt.Println("err157", err)
		return utils.BYTE_FALSE, errors.NewErr("[FS Govern] FsFileProve fileInfo serialize error!")
	}
	log.Debugf("filmInfo.PaidCount %v\n", filmInfo.PaidCount)
	utils.PutBytes(native, userFilmKeys, bf.Bytes())

	if filmInfo.Price == 0 {
		fmt.Println("add to buy list")
		if err := addFilmToUserBuyList(native, user, filmInfo); err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] Film added to buyer list failed err!")
		}

		fmt.Println("add to profit list list")
		if err := addFilmToOwnerProfitList(native, user, filmInfo); err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] Film added to owner list failed err!")
		}
		fmt.Println("done++++")
		return utils.BYTE_TRUE, nil
	}

	fmt.Println("make transfer")
	// transfer asset from user to owner
	if err := appCallTransfer(native, utils.UsdtContractAddress, user, owner, filmInfo.Price); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] Film transfer asset faileds err!")
	}

	fmt.Println("add to buy list")
	if err := addFilmToUserBuyList(native, user, filmInfo); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] Film added to buyer list failed err!")
	}

	fmt.Println("add to profit list list")
	if err := addFilmToOwnerProfitList(native, user, filmInfo); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] Film added to owner list failed err!")
	}

	return utils.BYTE_TRUE, nil
}

func GetUserFilmList(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	buf, err := utils.DecodeBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] GetAllFilmList deserialize error!")
	}
	params := make([]interface{}, 0)
	err = json.Unmarshal(buf, &params)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] GetAllFilmList unmarshal error!")
	}
	if len(params) == 0 {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] GetAllFilmList params miss !")
	}
	user, err := common.AddressFromBase58(getStringValue(params[0]))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] GetAllFilmList AddressFromBase58 error!")
	}
	hashes, err := getUserFilmList(native, user)
	if err != nil {
		return nil, err
	}

	var list []*FilmInfo
	for _, hash := range hashes {
		item, err := getFilmInfo(native, user, hash)
		if err != nil {
			continue
		}

		filmInfo := &FilmInfo{}
		r := bytes.NewReader(item)
		err = filmInfo.Deserialize(r)
		if err != nil {
			continue
		}

		list = append(list, filmInfo)
	}
	return json.Marshal(list)
}

func GetAllFilmList(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	buf, err := utils.DecodeBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] GetAllFilmList deserialize error!")
	}
	params := make([]interface{}, 0)
	err = json.Unmarshal(buf, &params)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] GetAllFilmList unmarshal error!")
	}
	if len(params) == 0 {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] GetAllFilmList params miss !")
	}
	requestWallet, err := common.AddressFromBase58(getStringValue(params[0]))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] GetAllFilmList AddressFromBase58 error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	filmCountKey := GetFilmCountKey(contract)
	totalCount, err := utils.GetStorageUInt64(native, filmCountKey)
	if err != nil {
		return nil, errors.NewErr("[FS Profit] GetAllFilmList GetStorageUInt64 error!")
	}

	userDownloadedList, _ := getUserBuyHashesList(native, requestWallet)
	userDownloadedListMap := make(map[string]struct{}, 0)
	for _, hash := range userDownloadedList {
		userDownloadedListMap[string(hash)] = struct{}{}
	}

	var list []*FilmInfo
	for i := uint64(1); i <= totalCount; i++ {
		filmSearchKey := GetFilmKeyAtList(contract, i)
		filmSearchValue, err := utils.GetStorageItem(native, filmSearchKey)
		if err != nil {
			return nil, errors.NewErr("[FILM Govern] GetAllFilmList unmarshal error!")
		}
		if filmSearchValue == nil {
			continue
		}
		item, err := utils.GetStorageItem(native, filmSearchValue.Value)
		if err != nil {
			continue
		}
		if item == nil || len(item.Value) == 0 {
			continue
		}
		filmInfo := &FilmInfo{}
		r := bytes.NewReader(item.Value)
		err = filmInfo.Deserialize(r)
		if err != nil {
			continue
		}

		if filmInfo.Price == 0 || filmInfo.Owner.ToBase58() == requestWallet.ToBase58() {
			list = append(list, filmInfo)
			continue
		}

		// check if user has purchased
		if len(userDownloadedList) == 0 {
			filmInfo.Url = []byte("")
		} else {
			if _, ok := userDownloadedListMap[string(filmInfo.Hash)]; !ok {
				filmInfo.Url = []byte("")
			}
		}
		list = append(list, filmInfo)
	}
	return json.Marshal(list)
}

func GetFilmInfo(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	buf, err := utils.DecodeBytes(source)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FILM Govern] FilmPublish deserialize error!")
	}
	params := make([]interface{}, 0)
	err = json.Unmarshal(buf, &params)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FILM Govern] FilmPublish unmarshal error!")
	}
	fileHash := []byte(getStringValue(params[0]))
	owner, err := common.AddressFromBase58(getStringValue(params[1]))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] GetAllFilmList owner AddressFromBase58 error!")
	}
	item, err := getFilmInfo(native, owner, fileHash)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] GetAllFilmList  getFilmInfo error!")
	}
	filmInfo := &FilmInfo{}
	r := bytes.NewReader(item)
	err = filmInfo.Deserialize(r)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] Film Deserialize err!")
	}
	return json.Marshal(filmInfo)
}

func GetUserBuyRecordList(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	buf, err := utils.DecodeBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] GetAllFilmList deserialize error!")
	}
	params := make([]interface{}, 0)
	err = json.Unmarshal(buf, &params)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] GetAllFilmList unmarshal error!")
	}
	if len(params) == 0 {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] GetAllFilmList params miss !")
	}
	user, err := common.AddressFromBase58(getStringValue(params[0]))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] GetAllFilmList AddressFromBase58 error!")
	}
	l, err := getUserBuyRecordList(native, user)
	if err != nil {
		return nil, err
	}

	list := make([]BuyRecord, 0)
	for _, hash := range l.TxHashes {
		item, err := getUserBuyRecord(native, user, hash)
		if err != nil {
			continue
		}

		list = append(list, item)
	}
	return json.Marshal(list)
}

func GetUserProfitRecordList(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	buf, err := utils.DecodeBytes(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] GetAllFilmList deserialize error!")
	}
	params := make([]interface{}, 0)
	err = json.Unmarshal(buf, &params)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] GetAllFilmList unmarshal error!")
	}
	if len(params) == 0 {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] GetAllFilmList params miss !")
	}
	user, err := common.AddressFromBase58(getStringValue(params[0]))
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[FILM Govern] GetAllFilmList AddressFromBase58 error!")
	}
	l, err := getUserProfitList(native, user)
	fmt.Println("l", l)
	if err != nil {
		return nil, err
	}

	list := make([]ProfitRecord, 0)
	for _, hash := range l.TxHashes {
		item, err := getUserProfitRecord(native, user, hash)
		if err != nil {
			continue
		}

		list = append(list, item)
	}
	return json.Marshal(list)
}

func getFilmInfo(native *native.NativeService, owner common.Address, hash []byte) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	filmInfoKey := GenFilmInfoKey(contract, owner, hash)
	item, err := utils.GetStorageItem(native, filmInfoKey)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Profit] FsFileInfo GetStorageItem error!")
	}
	if item == nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Profit] FsFileInfo not found!")
	}
	return item.Value, nil
}

func getFileHashFromUrl(native *native.NativeService, url string) ([]byte, error) {
	strs := strings.Split(url, "://")
	namekey := dns.GenNameInfoKey(utils.OntDNSAddress, []byte(strs[0]), []byte(strs[1]))
	nameItem, err := utils.GetStorageItem(native, namekey)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FILM Govern] get fileInfo error!")
	}
	if nameItem == nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FILM Govern] get fileInfo error!")
	}
	var name dns.NameInfo
	infoReader := bytes.NewReader(nameItem.Value)
	err = name.Deserialize(infoReader)
	if err != nil {
		return nil, err
	}
	link := string(name.Name)
	idx := strings.Index(link, fmt.Sprintf("%s", "Qm"))
	if idx != -1 {
		return []byte(link[idx : idx+46]), nil
	}
	idx = strings.Index(link, fmt.Sprintf("%s", "zb"))
	if idx != -1 {
		return []byte(link[idx : idx+49]), nil
	}
	return nil, nil
}

// getFileInfo. get file info from fs contract with file hash
func getFileInfo(native *native.NativeService, fileHash []byte) (*savefs.FileInfo, error) {
	fileInfoKey := savefs.GenFsFileInfoKey(utils.OntFSContractAddress, fileHash)
	item, err := utils.GetStorageItem(native, fileInfoKey)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Profit] FsFileInfo GetStorageItem error!")
	}
	if item == nil {
		return nil, errors.NewErr("item is nil")
	}
	var fsFileInfo savefs.FileInfo
	fsFileInfoSource := common.NewZeroCopySource(item.Value)
	err = fsFileInfo.Deserialization(fsFileInfoSource)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Profit] FsFileInfo deserialize error!")
	}

	return &fsFileInfo, nil
}

// getUserFilmList. get user film list with user wallet address
func getUserFilmList(native *native.NativeService, user common.Address) ([][]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	userFilmKeys := GenUserFilmListKey(contract, user)
	item, err := utils.GetStorageItem(native, userFilmKeys)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Profit] FsFileInfo GetStorageItem error!")
	}
	if item == nil {
		return nil, nil
	}
	var userList UserFilmList
	source := common.NewZeroCopySource(item.Value)
	err = userList.Deserialization(source)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Profit] FsFileInfo deserialize error!")
	}
	return userList.FilmHashes, nil
}

// getUserBuyList. get user film list with user wallet address
func getUserBuyHashesList(native *native.NativeService, user common.Address) ([][]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	userDownloadFilmListKeys := GenUserFilmBuyListKey(contract, user)
	item, err := utils.GetStorageItem(native, userDownloadFilmListKeys)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Profit] FsFileInfo GetStorageItem error!")
	}
	if item == nil {
		return nil, nil
	}
	var downloadedList BuyRecordList
	source := common.NewZeroCopySource(item.Value)
	err = downloadedList.Deserialization(source)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Profit] FsFileInfo deserialize error!")
	}

	fileHashes := make([][]byte, 0)

	for _, id := range downloadedList.TxHashes {
		userDownloadFilmKey := GenUserFilmBuyInfoKey(contract, user, id)
		item, err := utils.GetStorageItem(native, userDownloadFilmKey)
		if err != nil || item == nil {
			continue
		}
		var downloaded BuyRecord
		source := common.NewZeroCopySource(item.Value)
		err = downloaded.Deserialization(source)
		if err != nil {
			continue
		}
		fileHashes = append(fileHashes, downloaded.FilmHash)
	}

	return fileHashes, nil
}

// addFilmToUserFilmList. add film to user's film list
func addFilmToUserFilmList(native *native.NativeService, owner common.Address, newFilmHash []byte) error {
	userFilmList, err := getUserFilmList(native, owner)
	if err != nil {
		log.Errorf("get user film err  err %s", err)
		return errors.NewErr("[FILM Govern] Get user film list failed!")
	}
	log.Debugf("userFilmList %v\n", len(userFilmList))
	if userFilmList == nil {
		userFilmList = make([][]byte, 0)
	}
	for _, film := range userFilmList {
		if bytes.Equal(film, newFilmHash) {
			log.Debugf("alread exists+++")
			return nil
		}
	}
	userFilmList = append(userFilmList, newFilmHash)
	userFilm := &UserFilmList{
		Owner:      owner,
		Num:        uint64(len(userFilmList)),
		FilmHashes: userFilmList,
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	userFilmKeys := GenUserFilmListKey(contract, owner)
	bf := new(bytes.Buffer)
	if err = userFilm.Serialize(bf); err != nil {
		return errors.NewErr("[FS Govern] FsFileProve fileInfo serialize error!")
	}
	utils.PutBytes(native, userFilmKeys, bf.Bytes())
	return nil
}

// addFilmToAllFilmList. add film to film list if not exists
func addFilmToAllFilmList(native *native.NativeService, film *FilmInfo, filmOwner common.Address) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	filmCountKey := GetFilmCountKey(contract)
	filmIndex, err := utils.GetStorageUInt64(native, filmCountKey)
	if err != nil {
		return errors.NewErr("[FS Profit] FsFileInfo GetStorageUInt64 error!")
	}
	filmIndex = filmIndex + 1
	filmKeyAtList := GetFilmKeyAtList(contract, filmIndex)
	filmInfoKey := GenFilmInfoKey(contract, filmOwner, film.Hash)

	// filmKeyValueAtList := fmt.Sprintf("hash=%s&owner=%s&type=%d&year=%d&available=%t&region=%s&name=%s",
	// 	film.Hash, filmOwner.ToBase58(), film.Type, film.ReleaseYear, film.Available,
	// 	hex.EncodeToString(film.Region), hex.EncodeToString(film.Name))
	log.Debugf("add film to all film list key: %s, value: %v\n", filmKeyAtList, filmInfoKey)
	utils.PutBytes(native, filmKeyAtList, []byte(filmInfoKey))
	newFilmCountItem := utils.GenUInt64StorageItem(filmIndex)
	utils.PutBytes(native, filmCountKey, newFilmCountItem.Value)
	return nil
}

// addFilmToUserFilmList. add film to user's film list
func addFilmToUserBuyList(native *native.NativeService, user common.Address, filmInfo *FilmInfo) error {
	buyList, err := getUserBuyRecordList(native, user)
	if err != nil {
		log.Errorf("get user film err  err %s", err)
		return errors.NewErr("[FILM Govern] Get user film list failed!")
	}
	log.Debugf("buyList %v\n", len(buyList.TxHashes))
	if buyList.TxHashes == nil {
		buyList.TxHashes = make([][]byte, 0)
	}

	for _, hash := range buyList.TxHashes {
		item, err := getUserBuyRecord(native, user, hash)
		if err != nil {
			continue
		}
		if string(item.FilmHash) == string(filmInfo.Hash) {
			// already purchased
			return nil
		}
	}

	txHash := native.Tx.Hash()
	buyList.TxHashes = append(buyList.TxHashes, txHash[:])
	buyList.RecordNum++
	buyList.Owner = user

	var r BuyRecord
	r.TxHash = txHash[:]
	r.FilmHash = filmInfo.Hash
	r.BuyAt = uint64(native.Time)
	r.FilmOwner = filmInfo.Owner
	r.Cost = filmInfo.Price
	r.BlockHeight = uint64(native.Height)

	contract := native.ContextRef.CurrentContext().ContractAddress
	userBuyListKeys := GenUserFilmBuyListKey(contract, user)
	bf := new(bytes.Buffer)
	if err = buyList.Serialize(bf); err != nil {
		return errors.NewErr("[FS Govern] FsFileProve buyList serialize error!")
	}
	utils.PutBytes(native, userBuyListKeys, bf.Bytes())

	bf = new(bytes.Buffer)
	if err = r.Serialize(bf); err != nil {
		return errors.NewErr("[FS Govern] FsFileProve BuyRecord serialize error!")
	}
	userBuyRecordKey := GenUserFilmBuyInfoKey(contract, user, txHash[:])
	utils.PutBytes(native, userBuyRecordKey, bf.Bytes())
	return nil
}

// addFilmToOwnerProfitList. add film to owner's profit list
func addFilmToOwnerProfitList(native *native.NativeService, buyer common.Address, filmInfo *FilmInfo) error {
	profitList, err := getUserProfitList(native, filmInfo.Owner)
	if err != nil {
		log.Errorf("get user film err  err %s", err)
		return errors.NewErr("[FILM Govern] Get user film list failed!")
	}
	log.Debugf("addFilmToOwnerProfitList %v\n", profitList.TxHashes)
	if profitList.TxHashes == nil {
		profitList.TxHashes = make([][]byte, 0)
	}

	for _, hash := range profitList.TxHashes {
		item, err := getUserProfitRecord(native, filmInfo.Owner, hash)
		if err != nil {
			continue
		}
		if string(item.FilmHash) == string(filmInfo.Hash) {
			// already purchased
			fmt.Println("already buy", item)
			return nil
		}
	}

	txHash := native.Tx.Hash()
	profitList.TxHashes = append(profitList.TxHashes, txHash[:])
	profitList.Num++
	profitList.Owner = filmInfo.Owner

	log.Debugf("GenUserFilmProfitListKey %v\n", len(profitList.TxHashes))

	var r ProfitRecord
	r.TxHash = txHash[:]
	r.FilmHash = filmInfo.Hash
	r.BuyAt = uint64(native.Time)
	r.Payer = buyer
	r.PayAmount = filmInfo.Price
	r.BlockHeight = uint64(native.Height)

	contract := native.ContextRef.CurrentContext().ContractAddress
	profitListKeys := GenUserFilmProfitListKey(contract, filmInfo.Owner)
	bf := new(bytes.Buffer)
	if err = profitList.Serialize(bf); err != nil {
		return errors.NewErr("[FS Govern] FsFileProve buyList serialize error!")
	}
	utils.PutBytes(native, profitListKeys, bf.Bytes())

	bf = new(bytes.Buffer)
	if err = r.Serialize(bf); err != nil {
		return errors.NewErr("[FS Govern] FsFileProve BuyRecord serialize error!")
	}
	profitRecordKey := GenUserFilmProfitInfoKey(contract, filmInfo.Owner, txHash[:])
	utils.PutBytes(native, profitRecordKey, bf.Bytes())
	return nil
}

// getUserBuyRecordList. get user film list with user wallet address
func getUserBuyRecordList(native *native.NativeService, user common.Address) (BuyRecordList, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	userDownloadFilmListKeys := GenUserFilmBuyListKey(contract, user)
	item, err := utils.GetStorageItem(native, userDownloadFilmListKeys)
	var list BuyRecordList
	if err != nil {
		return list, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Profit] FsFileInfo GetStorageItem error!")
	}
	if item == nil {
		return list, nil
	}
	source := common.NewZeroCopySource(item.Value)
	err = list.Deserialization(source)
	if err != nil {
		return list, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Profit] FsFileInfo deserialize error!")
	}

	return list, nil
}

// getUserBuyRecordList. get user film list with user wallet address
func getUserBuyRecord(native *native.NativeService, user common.Address, txHash []byte) (BuyRecord, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	key := GenUserFilmBuyInfoKey(contract, user, txHash)
	item, err := utils.GetStorageItem(native, key)
	var list BuyRecord
	if err != nil {
		return list, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Profit] FsFileInfo GetStorageItem error!")
	}
	if item == nil {
		return list, nil
	}
	source := common.NewZeroCopySource(item.Value)
	err = list.Deserialization(source)
	if err != nil {
		return list, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Profit] FsFileInfo deserialize error!")
	}

	return list, nil
}

// getUserProfitList. get user buy record item with user wallet address
func getUserProfitList(native *native.NativeService, user common.Address) (ProfitRecordList, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	profitRecordKey := GenUserFilmProfitListKey(contract, user)
	fmt.Println("profitRecordKey", profitRecordKey, user.ToBase58())
	item, err := utils.GetStorageItem(native, profitRecordKey)
	var list ProfitRecordList
	if err != nil {
		fmt.Println("err742", err)
		return list, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Profit] ProfitRecordList GetStorageItem error!")
	}
	if item == nil {
		fmt.Println("err746", err)
		return list, nil
	}
	source := common.NewZeroCopySource(item.Value)
	err = list.Deserialization(source)
	if err != nil {
		return list, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Profit] ProfitRecordList deserialize error!")
	}

	return list, nil
}

// getUserProfitRecord. get user profit item with user wallet address
func getUserProfitRecord(native *native.NativeService, user common.Address, txHash []byte) (ProfitRecord, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	key := GenUserFilmProfitInfoKey(contract, user, txHash)
	item, err := utils.GetStorageItem(native, key)
	var list ProfitRecord
	if err != nil {
		return list, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Profit] FsFileInfo GetStorageItem error!")
	}
	if item == nil {
		return list, nil
	}
	source := common.NewZeroCopySource(item.Value)
	err = list.Deserialization(source)
	if err != nil {
		return list, errors.NewDetailErr(err, errors.ErrNoCode, "[FS Profit] FsFileInfo deserialize error!")
	}

	return list, nil
}
