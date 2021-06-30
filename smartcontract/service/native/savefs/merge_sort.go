package savefs

import (
	"github.com/saveio/themis/common"
	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/service/native"
	"strings"
)

func merge(l []*SectorFileInfo, r []*SectorFileInfo) []*SectorFileInfo {
	result := make([]*SectorFileInfo, 0)

	for len(l) > 0 || len(r) > 0 {
		if len(l) > 0 && len(r) > 0 {
			if strings.Compare(string(l[0].FileHash), string(r[0].FileHash)) < 0 {
				result = append(result, l[0])
				l = l[1:]
			} else {
				result = append(result, r[0])
				r = r[1:]
			}
		} else if len(l) > 0 {
			result = append(result, l[0])
			l = l[1:]
		} else {
			result = append(result, r[0])
			r = r[1:]
		}
	}

	return result
}

/*
func mergeSort(list []*SectorFileInfo) []*SectorFileInfo {
	switch len(list) {
	case 0:
		return list
	case 1:
		return list
	default:
		m := len(list) / 2
		l := list[:m]
		r := list[m:]
		return merge(mergeSort(l), mergeSort(r))
	}
}
*/
func mergeSortSectorFileInfo(native *native.NativeService, nodeAddr common.Address, sectorID uint64) ([]*SectorFileInfo, error) {
	groupNum, err := getSectorFileInfoGroupNum(native, nodeAddr, sectorID)
	if err != nil {
		return nil, errors.NewErr("getSectorFileInfoGroupNum error!")
	}

	if groupNum == 0 {
		return nil, errors.NewErr("group num is 0!")
	}

	if groupNum == 1 {
		group, err := getSectorFileInfoGroup(native, nodeAddr, sectorID, 1)
		if err != nil {
			return nil, errors.NewErr("getSectorFileInfoGroup error !")
		}
		return group.FileList, nil
	}

	sectorFileInfoMap := make(map[uint64][]*SectorFileInfo)

	for len(sectorFileInfoMap) != 1 {
		var loops int
		var err error
		var group1 *SectorFileInfoGroup
		var group2 *SectorFileInfoGroup
		var fileInfos1 []*SectorFileInfo
		var fileInfos2 []*SectorFileInfo

		firstLoop := false
		mapLen := len(sectorFileInfoMap)

		if mapLen == 0 {
			loops = int(groupNum+1) / 2
			mapLen = int(groupNum)
			firstLoop = true
		} else {
			loops = (len(sectorFileInfoMap) + 1) / 2
		}

		for i := uint64(0); i < uint64(loops); i++ {
			// only read from db in first loop, following loops will read directly from map
			if firstLoop {
				// note here id starts from 1 not 0
				group1, err = getSectorFileInfoGroup(native, nodeAddr, sectorID, 2*i+1)
				if err != nil {
					return nil, errors.NewErr("getSectorFileInfoGroup error !")
				}
				fileInfos1 = group1.FileList
			} else {
				fileInfos1 = sectorFileInfoMap[2*i]
			}

			// if one group left just insert to map, here 2*i +1 is the index, not id
			if 2*i+1 == uint64(mapLen) {
				sectorFileInfoMap[i] = fileInfos1
				if i != 0 {
					delete(sectorFileInfoMap, 2*i)
				}
				break
			}

			if firstLoop {
				group2, err = getSectorFileInfoGroup(native, nodeAddr, sectorID, 2*i+2)
				if err != nil {
					return nil, errors.NewErr("getSectorFileInfoGroup error !")
				}
				fileInfos2 = group2.FileList
				// check min and max of two groups to see if already sorted
				if strings.Compare(string(group1.MaxFileHash), string(group2.MinFileHash)) < 0 {
					sectorFileInfoMap[i] = append(group1.FileList, group2.FileList...)
				} else {
					sectorFileInfoMap[i] = merge(group1.FileList, group2.FileList)
				}
			} else {
				fileInfos2 = sectorFileInfoMap[2*i+1]
				sectorFileInfoMap[i] = merge(fileInfos1, fileInfos2)
			}

			// delete keys that no long used
			if i != 0 {
				delete(sectorFileInfoMap, 2*i)
				delete(sectorFileInfoMap, 2*i+1)
			} else {
				delete(sectorFileInfoMap, 2*i+1)
			}
		}
	}

	return sectorFileInfoMap[0], nil
}
