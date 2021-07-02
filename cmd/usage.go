/*
 * Copyright (C) 2019 The themis Authors
 * This file is part of The themis library.
 *
 * The themis is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The themis is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The themis.  If not, see <http://www.gnu.org/licenses/>.
 */
package cmd

import (
	"encoding/hex"
	"fmt"
	"strings"

	cmdcom "github.com/saveio/themis/cmd/common"
	"github.com/saveio/themis/cmd/utils"
	"github.com/saveio/themis/common"
	nutils "github.com/saveio/themis/smartcontract/service/native/utils"
	"github.com/urfave/cli"
)

var SendTxCommand = cli.Command{
	Name:        "sendtx",
	Usage:       "Send raw transaction to Ontology",
	Description: "Send raw transaction to Ontology.",
	ArgsUsage:   "<rawtx>",
	Action:      sendTx,
	Flags: []cli.Flag{
		utils.RPCPortFlag,
		utils.PrepareExecTransactionFlag,
	},
}

func sendTx(ctx *cli.Context) error {
	SetRpcPort(ctx)
	if ctx.NArg() < 1 {
		PrintErrorMsg("Missing raw tx argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	rawTx := ctx.Args().First()

	isPre := ctx.IsSet(utils.GetFlagName(utils.PrepareExecTransactionFlag))
	if isPre {
		preResult, err := utils.PrepareSendRawTransaction(rawTx)
		if err != nil {
			return err
		}
		if preResult.State == 0 {
			return fmt.Errorf("prepare execute transaction failed. %v", preResult)
		}
		PrintInfoMsg("Prepare execute transaction success.")
		PrintInfoMsg("Gas limit:%d", preResult.Gas)
		PrintInfoMsg("Result:%v", preResult.Result)
		return nil
	}
	txHash, err := utils.SendRawTransactionData(rawTx)
	if err != nil {
		return err
	}
	PrintInfoMsg("Send transaction success.")
	PrintInfoMsg("  TxHash:%s", txHash)
	PrintInfoMsg("\nTip:")
	PrintInfoMsg("  Using './themis info status %s' to query transaction status.", txHash)
	return nil
}

var TxCommond = cli.Command{
	Name:  "buildtx",
	Usage: "Build transaction",
	Subcommands: []cli.Command{
		TransferTxCommond,
		ApproveTxCommond,
		TransferFromTxCommond,
		ApproveCandidateTxCommond,
		UpdateConfigTxCommond,
	},
	Description: "Build transaction",
}

var TransferTxCommond = cli.Command{
	Name:        "transfer",
	Usage:       "Build transfer transaction",
	Description: "Build transfer transaction.",
	Action:      transferTx,
	Flags: []cli.Flag{
		utils.WalletFileFlag,
		utils.TransactionGasPriceFlag,
		utils.TransactionGasLimitFlag,
		utils.TransactionPayerFlag,
		utils.TransactionAssetFlag,
		utils.TransactionFromFlag,
		utils.TransactionToFlag,
		utils.TransactionAmountFlag,
	},
}

var ApproveTxCommond = cli.Command{
	Name:        "approve",
	Usage:       "Build approve transaction",
	Description: "Build approve transaction.",
	Action:      approveTx,
	Flags: []cli.Flag{
		utils.WalletFileFlag,
		utils.TransactionGasPriceFlag,
		utils.TransactionGasLimitFlag,
		utils.TransactionPayerFlag,
		utils.ApproveAssetFlag,
		utils.ApproveAssetFromFlag,
		utils.ApproveAssetToFlag,
		utils.ApproveAmountFlag,
	},
}

var TransferFromTxCommond = cli.Command{
	Name:        "transferfrom",
	Usage:       "Build transfer from transaction",
	Description: "Build transfer from transaction.",
	Action:      transferFromTx,
	Flags: []cli.Flag{
		utils.WalletFileFlag,
		utils.TransactionGasPriceFlag,
		utils.TransactionGasLimitFlag,
		utils.ApproveAssetFlag,
		utils.TransactionPayerFlag,
		utils.TransferFromSenderFlag,
		utils.ApproveAssetFromFlag,
		utils.ApproveAssetToFlag,
		utils.TransferFromAmountFlag,
	},
}

var ApproveCandidateTxCommond = cli.Command{
	Name:        "approvecand",
	Usage:       "Build approve consensus,dns candidate transaction",
	Description: "Build approve consensus,dns candidate transaction.",
	Action:      approveCandidateTx,
	Flags: []cli.Flag{
		utils.WalletFileFlag,
		utils.TransactionGasPriceFlag,
		utils.TransactionGasLimitFlag,
		utils.TransactionPayerFlag,
		utils.RejectCandidateFlag,
		utils.ApproveCandidatePubkeyFlag,
		utils.ApproveCandidateRoleFlag,
	},
}

var UpdateConfigTxCommond = cli.Command{
	Name:        "updateconfig",
	Usage:       "Update consensus config in governance contract",
	Description: "Update consensus config in governance contract.",
	Action:      updateConfigTx,
	Flags: []cli.Flag{
		utils.WalletFileFlag,
		utils.TransactionGasPriceFlag,
		utils.TransactionGasLimitFlag,
		utils.TransactionPayerFlag,
		utils.ConfigNFlag,
		utils.ConfigCFlag,
		utils.ConfigKFlag,
		utils.ConfigLFlag,
		utils.ConfigBlockMsgDelayFlag,
		utils.ConfigHashMsgDelayFlag,
		utils.ConfigPeerHandshakeTimeoutFlag,
		utils.ConfigMaxBlockChangeViewFlag,
	},
}

func transferTx(ctx *cli.Context) error {
	if !ctx.IsSet(utils.GetFlagName(utils.TransactionToFlag)) ||
		!ctx.IsSet(utils.GetFlagName(utils.TransactionFromFlag)) ||
		!ctx.IsSet(utils.GetFlagName(utils.TransactionAmountFlag)) {
		PrintErrorMsg("Missing %s %s or %s argument.", utils.TransactionToFlag.Name, utils.TransactionFromFlag.Name, utils.TransactionAmountFlag.Name)
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	gasPrice := ctx.Uint64(utils.TransactionGasPriceFlag.Name)
	gasLimit := ctx.Uint64(utils.TransactionGasLimitFlag.Name)

	asset := ctx.String(utils.GetFlagName(utils.TransactionAssetFlag))
	if asset == "" {
		asset = utils.ASSET_USDT
	}
	from := ctx.String(utils.GetFlagName(utils.TransactionFromFlag))
	fromAddr, err := cmdcom.ParseAddress(from, ctx)
	if err != nil {
		return err
	}
	to := ctx.String(utils.GetFlagName(utils.TransactionToFlag))
	toAddr, err := cmdcom.ParseAddress(to, ctx)
	if err != nil {
		return err
	}

	var payer common.Address
	payerAddr := ctx.String(utils.GetFlagName(utils.TransactionPayerFlag))
	if payerAddr != "" {
		payerAddr, err = cmdcom.ParseAddress(payerAddr, ctx)
		if err != nil {
			return err
		}
	} else {
		payerAddr = fromAddr
	}

	payer, err = common.AddressFromBase58(payerAddr)
	if err != nil {
		return fmt.Errorf("invalid payer address:%s", err)
	}

	var amount uint64
	amountStr := ctx.String(utils.TransactionAmountFlag.Name)
	switch strings.ToLower(asset) {
	case utils.ASSET_USDT:
		amount = utils.ParseUsdt(amountStr)
		amountStr = utils.FormatUsdt(amount)
	default:
		return fmt.Errorf("unsupport asset:%s", asset)
	}

	err = utils.CheckAssetAmount(asset, amount)
	if err != nil {
		return err
	}

	mutTx, err := utils.TransferTx(gasPrice, gasLimit, asset, fromAddr, toAddr, amount)
	if err != nil {
		return err
	}
	mutTx.Payer = payer

	tx, err := mutTx.IntoImmutable()
	if err != nil {
		return fmt.Errorf("IntoImmutable error:%s", err)
	}
	sink := common.ZeroCopySink{}
	tx.Serialization(&sink)
	PrintInfoMsg("Transfer raw tx:")
	PrintInfoMsg(hex.EncodeToString(sink.Bytes()))
	return nil
}

func approveTx(ctx *cli.Context) error {
	asset := ctx.String(utils.GetFlagName(utils.ApproveAssetFlag))
	from := ctx.String(utils.GetFlagName(utils.ApproveAssetFromFlag))
	to := ctx.String(utils.GetFlagName(utils.ApproveAssetToFlag))
	amountStr := ctx.String(utils.GetFlagName(utils.ApproveAmountFlag))
	if asset == "" ||
		from == "" ||
		to == "" ||
		amountStr == "" {
		PrintErrorMsg("Missing %s %s %s or %s argument.", utils.ApproveAssetFlag.Name, utils.ApproveAssetFromFlag.Name, utils.ApproveAssetToFlag.Name, utils.ApproveAmountFlag.Name)
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	fromAddr, err := cmdcom.ParseAddress(from, ctx)
	if err != nil {
		return err
	}
	toAddr, err := cmdcom.ParseAddress(to, ctx)
	if err != nil {
		return err
	}

	var payer common.Address
	payerAddr := ctx.String(utils.GetFlagName(utils.TransactionPayerFlag))
	if payerAddr != "" {
		payerAddr, err = cmdcom.ParseAddress(payerAddr, ctx)
		if err != nil {
			return err
		}
	} else {
		payerAddr = fromAddr
	}

	payer, err = common.AddressFromBase58(payerAddr)
	if err != nil {
		return fmt.Errorf("invalid payer address:%s", err)
	}

	var amount uint64
	switch strings.ToLower(asset) {
	case utils.ASSET_USDT:
		amount = utils.ParseUsdt(amountStr)
		amountStr = utils.FormatUsdt(amount)
	default:
		return fmt.Errorf("unsupport asset:%s", asset)
	}

	err = utils.CheckAssetAmount(asset, amount)
	if err != nil {
		return err
	}

	gasPrice := ctx.Uint64(utils.TransactionGasPriceFlag.Name)
	gasLimit := ctx.Uint64(utils.TransactionGasLimitFlag.Name)

	mutTx, err := utils.ApproveTx(gasPrice, gasLimit, asset, fromAddr, toAddr, amount)
	if err != nil {
		return err
	}
	mutTx.Payer = payer

	tx, err := mutTx.IntoImmutable()
	if err != nil {
		return fmt.Errorf("IntoImmutable error:%s", err)
	}
	sink := common.ZeroCopySink{}
	tx.Serialization(&sink)
	PrintInfoMsg("Approve raw tx:")
	PrintInfoMsg(hex.EncodeToString(sink.Bytes()))
	return nil
}

func transferFromTx(ctx *cli.Context) error {
	asset := ctx.String(utils.GetFlagName(utils.ApproveAssetFlag))
	from := ctx.String(utils.GetFlagName(utils.ApproveAssetFromFlag))
	to := ctx.String(utils.GetFlagName(utils.ApproveAssetToFlag))
	amountStr := ctx.String(utils.GetFlagName(utils.TransferFromAmountFlag))
	if asset == "" ||
		from == "" ||
		to == "" ||
		amountStr == "" {
		PrintErrorMsg("Missing %s %s %s or %s argument.", utils.ApproveAssetFlag.Name, utils.ApproveAssetFromFlag.Name, utils.ApproveAssetToFlag.Name, utils.TransferFromAmountFlag.Name)
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	fromAddr, err := cmdcom.ParseAddress(from, ctx)
	if err != nil {
		return err
	}
	toAddr, err := cmdcom.ParseAddress(to, ctx)
	if err != nil {
		return err
	}

	var sendAddr string
	sender := ctx.String(utils.GetFlagName(utils.TransferFromSenderFlag))
	if sender == "" {
		sendAddr = toAddr
	} else {
		sendAddr, err = cmdcom.ParseAddress(sender, ctx)
		if err != nil {
			return err
		}
	}

	var payer common.Address
	payerAddr := ctx.String(utils.GetFlagName(utils.TransactionPayerFlag))
	if payerAddr != "" {
		payerAddr, err = cmdcom.ParseAddress(payerAddr, ctx)
		if err != nil {
			return err
		}
	} else {
		payerAddr = sendAddr
	}

	payer, err = common.AddressFromBase58(payerAddr)
	if err != nil {
		return fmt.Errorf("invalid payer address:%s", err)
	}

	var amount uint64
	switch strings.ToLower(asset) {
	case utils.ASSET_USDT:
		amount = utils.ParseUsdt(amountStr)
		amountStr = utils.FormatUsdt(amount)
	default:
		return fmt.Errorf("unsupport asset:%s", asset)
	}

	err = utils.CheckAssetAmount(asset, amount)
	if err != nil {
		return err
	}

	gasPrice := ctx.Uint64(utils.TransactionGasPriceFlag.Name)
	gasLimit := ctx.Uint64(utils.TransactionGasLimitFlag.Name)

	mutTx, err := utils.TransferFromTx(gasPrice, gasLimit, asset, sendAddr, fromAddr, toAddr, amount)
	if err != nil {
		return err
	}
	mutTx.Payer = payer

	tx, err := mutTx.IntoImmutable()
	if err != nil {
		return fmt.Errorf("IntoImmutable error:%s", err)
	}
	sink := common.ZeroCopySink{}
	tx.Serialization(&sink)
	PrintInfoMsg("TransferFrom raw tx:")
	PrintInfoMsg(hex.EncodeToString(sink.Bytes()))
	return nil
}

func approveCandidateTx(ctx *cli.Context) error {
	var err error

	peerPubkey := ctx.String(utils.GetFlagName(utils.ApproveCandidatePubkeyFlag))
	payerAddr := ctx.String(utils.GetFlagName(utils.TransactionPayerFlag))

	if peerPubkey == "" ||
		payerAddr == "" {
		PrintErrorMsg("Missing %s,%s argument.", utils.ApproveCandidatePubkeyFlag.Name, utils.TransactionPayerFlag.Name)
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	var payer common.Address
	payerAddr := ctx.String(utils.GetFlagName(utils.TransactionPayerFlag))
	if payerAddr != "" {
		payerAddr, err = cmdcom.ParseAddress(payerAddr, ctx)
		if err != nil {
			return err
		}
	} else {
		payerAddr = accAddr
	}
	payer, err = common.AddressFromBase58(payerAddr)
	if err != nil {
		return fmt.Errorf("invalid payer address:%s", err)
	}

	gasPrice := ctx.Uint64(utils.TransactionGasPriceFlag.Name)
	gasLimit := ctx.Uint64(utils.TransactionGasLimitFlag.Name)
	role := ctx.String(utils.GetFlagName(utils.ApproveCandidateRoleFlag))
	reject := ctx.Bool(utils.GetFlagName(utils.RejectCandidateFlag))

	mutTx, err := utils.ApproveCandidateTx(gasPrice, gasLimit, reject, peerPubkey, role)
	if err != nil {
		return err
	}

	mutTx.Payer = payer

	tx, err := mutTx.IntoImmutable()
	if err != nil {
		return fmt.Errorf("IntoImmutable error:%s", err)
	}
	sink := common.ZeroCopySink{}
	err = tx.Serialization(&sink)
	if err != nil {
		return fmt.Errorf("tx serialization error:%s", err)
	}
	if reject {
		PrintInfoMsg("Reject %s candidate raw tx:", role)
	} else {
		PrintInfoMsg("Approve %s candidate raw tx:", role)
	}
	PrintInfoMsg(hex.EncodeToString(sink.Bytes()))
	return nil
}

func updateConfigTx(ctx *cli.Context) error {
	var err error

	payerAddr := ctx.String(utils.GetFlagName(utils.TransactionPayerFlag))

	if payerAddr == "" {
		PrintErrorMsg("Missing %s argument.", utils.TransactionPayerFlag.Name)
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	if payerAddr != "" {
		payerAddr, err = cmdcom.ParseAddress(payerAddr, ctx)
		if err != nil {
			return err
		}
	}
	payer, err := common.AddressFromBase58(payerAddr)
	if err != nil {
		return fmt.Errorf("invalid payer address:%s", err)
	}

	gasPrice := ctx.Uint64(utils.TransactionGasPriceFlag.Name)
	gasLimit := ctx.Uint64(utils.TransactionGasLimitFlag.Name)

	configure := &gov.Configuration{N: uint32(ctx.Uint64(utils.ConfigNFlag.Name)),
		C:                    uint32(ctx.Uint64(utils.ConfigCFlag.Name)),
		K:                    uint32(ctx.Uint64(utils.ConfigKFlag.Name)),
		L:                    uint32(ctx.Uint64(utils.ConfigLFlag.Name)),
		BlockMsgDelay:        uint32(ctx.Uint64(utils.ConfigBlockMsgDelayFlag.Name)),
		HashMsgDelay:         uint32(ctx.Uint64(utils.ConfigHashMsgDelayFlag.Name)),
		PeerHandshakeTimeout: uint32(ctx.Uint64(utils.ConfigPeerHandshakeTimeoutFlag.Name)),
		MaxBlockChangeView:   uint32(ctx.Uint64(utils.ConfigMaxBlockChangeViewFlag.Name)),
	}

	invokeCode, err := cutils.BuildNativeInvokeCode(nutils.GovernanceContractAddress,
		utils.VERSION_TRANSACTION, gov.UPDATE_CONFIG, []interface{}{configure})

	if err != nil {
		return fmt.Errorf("build invoke code error:%s", err)
	}
	mutTx := utils.NewInvokeTransaction(gasPrice, gasLimit, invokeCode)

	mutTx.Payer = payer
	tx, err := mutTx.IntoImmutable()
	if err != nil {
		return fmt.Errorf("IntoImmutable error:%s", err)
	}
	sink := common.ZeroCopySink{}
	err = tx.Serialization(&sink)
	if err != nil {
		return fmt.Errorf("tx serialization error:%s", err)
	}

	PrintInfoMsg("Update config raw tx:")
	PrintInfoMsg(hex.EncodeToString(sink.Bytes()))
	return nil
}
