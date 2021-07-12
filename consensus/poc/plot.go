package poc

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/core/utils"
)

type Plot struct {
	AccountId  uint64
	StartNonce uint64
	Nonces     uint64
	FilePath   string
	FileHandle *os.File
	ReadOffset uint64
	Removed    bool
}

type PlotsDetail struct {
	Plots  []*Plot
	Lookup map[string]*Plot
}

func NewPlot(path string, useDirectIo bool) (*Plot, error) {

	parts := strings.Split(path, "_")
	if len(parts) < 3 {
		return nil, fmt.Errorf("plot file :s has wrong name format", path)
	}

	PthSep := string(os.PathSeparator)
	idStrs := strings.Split(parts[0], PthSep)

	accountId, _ := strconv.ParseUint(idStrs[len(idStrs)-1], 10, 64)
	startNonce, _ := strconv.ParseUint(parts[1], 10, 64)
	nonces, _ := strconv.ParseUint(parts[2], 10, 64)
	log.Debugf("accountId: %d, startNonce: %d, nonces: %d\n", accountId, startNonce, nonces)

	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("plot file:%s error:%s", path, err)
	}
	size := fileInfo.Size()
	expectedSize := int64(nonces * utils.NONCE_SIZE)
	if size != expectedSize {
		return nil, fmt.Errorf("plot file:%s has wrong size:%d, expected:%d", path, size, expectedSize)
	}

	if useDirectIo {
	} else {
	}

	plot := &Plot{
		AccountId:  accountId,
		StartNonce: startNonce,
		Nonces:     nonces,
		FilePath:   path,
		ReadOffset: 0,
	}

	return plot, nil
}
