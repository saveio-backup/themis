package dns

import (
	"bytes"
	"fmt"
	"io"

	"github.com/saveio/themis/errors"
	"github.com/saveio/themis/smartcontract/service/native"
	"github.com/saveio/themis/smartcontract/service/native/utils"
)

type Plugin struct {
	NameKey []byte
}

type PluginList struct {
	PluginNum uint64
	List      []Plugin
}

func (this *PluginList) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.PluginNum); err != nil {
		return fmt.Errorf("[PluginList] [PluginNum:%v] serialize from error:%v", this.PluginNum, err)
	}

	for index := 0; uint64(index) < this.PluginNum; index++ {
		if err := utils.WriteBytes(w, this.List[index].NameKey); err != nil {
			return fmt.Errorf("[PluginList] [PluginList:%v] serialize from error:%v", this.List[index].NameKey, err)
		}
	}
	return nil
}

func (this *PluginList) Deserialize(r io.Reader) error {
	var err error
	if this.PluginNum, err = utils.ReadVarUint(r); err != nil {
		return fmt.Errorf("[PluginList] [PluginNum] deserialize from error:%v", err)
	}
	var tmpNameKey []byte
	for index := 0; uint64(index) < this.PluginNum; index++ {
		if tmpNameKey, err = utils.ReadBytes(r); err != nil {
			return fmt.Errorf("[PluginList] [PluginList] deserialize from error:%v", err)
		}
		plugin := Plugin{tmpNameKey}
		this.List = append(this.List, plugin)
	}
	return nil
}

func (this *PluginList) Add(nameKey []byte) {
	flag := false
	for i := uint64(0); i < this.PluginNum; i++ {
		if bytes.Equal(this.List[i].NameKey, nameKey) {
			flag = true
			break
		}
	}
	if !flag {
		plugin := Plugin{nameKey}
		this.List = append(this.List, plugin)
		this.PluginNum++
	}
}

func (this *PluginList) Del(nameKey []byte) error {
	var i uint64
	if this.PluginNum == 0 {
		return nil
	}
	for i = 0; i < this.PluginNum; i++ {
		if bytes.Equal(this.List[i].NameKey, nameKey) {
			this.List = append(this.List[:i], this.List[i+1:]...)
			this.PluginNum -= 1
			break
		}
	}
	return nil
}

func AddPluginToList(native *native.NativeService, nameKey []byte) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	pluginListKey := GenPluginListKey(contract)
	return addPluginToList(native, pluginListKey, nameKey)
}

func DelPluginFromList(native *native.NativeService, nameKey []byte) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	pluginListKey := GenPluginListKey(contract)
	return delPluginToList(native, []byte(pluginListKey), nameKey)
}

func GetDnsPliginList(native *native.NativeService) (*PluginList, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	pluginListKey := GenPluginListKey(contract)
	return getDnsPluginList(native, pluginListKey)
}

func addPluginToList(native *native.NativeService, pluginListKey, nameKey []byte) error {
	var pluginList *PluginList
	pluginList, err := getDnsPluginList(native, pluginListKey)
	if pluginList == nil {
		pluginList = new(PluginList)
	}
	pluginList.Add(nameKey)
	pluginListBf := new(bytes.Buffer)
	err = pluginList.Serialize(pluginListBf)
	if err != nil {
		return errors.NewErr("[DNS PluginList] addPluginToList serialize error!")
	}
	utils.PutBytes(native, pluginListKey, pluginListBf.Bytes())
	return nil
}

func delPluginToList(native *native.NativeService, pluginListKey, nameKey []byte) error {
	var pluginList *PluginList
	pluginList, err := getDnsPluginList(native, pluginListKey)
	if err != nil {
		return errors.NewErr("[DNS PluginList] delPluginToList getDnsPluginList error!")
	}
	if pluginList == nil || pluginList.PluginNum == 0 {
		return nil
	}
	pluginList.Del(nameKey)
	pluginListBf := new(bytes.Buffer)
	err = pluginList.Serialize(pluginListBf)
	if err != nil {
		return errors.NewErr("[DNS PluginList] delPluginToList serialize error!")
	}
	utils.PutBytes(native, pluginListKey, pluginListBf.Bytes())
	return nil
}

func getDnsPluginList(native *native.NativeService, pluginListKey []byte) (*PluginList, error) {
	item, err := utils.GetStorageItem(native, pluginListKey)
	if err != nil {
		return nil, errors.NewErr("[DNS PluginList] DnsPluginList GetStorageItem error!")
	}
	if item == nil {
		return &PluginList{0, nil}, nil
	}
	var dnsPluginList PluginList
	reader := bytes.NewReader(item.Value)
	err = dnsPluginList.Deserialize(reader)
	if err != nil {
		return nil, errors.NewErr("[DNS PluginList] DnsPluginList deserialize error!")
	}
	return &dnsPluginList, nil
}
