package start

import (
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common/fdlimit"
	alog "github.com/ontio/ontology-eventbus/log"
	"github.com/saveio/themis/account"
	"github.com/saveio/themis/cmd"
	cmdcom "github.com/saveio/themis/cmd/common"
	"github.com/saveio/themis/cmd/utils"
	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/config"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/consensus"
	"github.com/saveio/themis/core/genesis"
	"github.com/saveio/themis/core/ledger"
	"github.com/saveio/themis/crypto/keypair"
	"github.com/saveio/themis/events"
	bactor "github.com/saveio/themis/http/base/actor"
	"github.com/saveio/themis/http/graphql"
	"github.com/saveio/themis/http/jsonrpc"
	"github.com/saveio/themis/http/localrpc"
	"github.com/saveio/themis/http/nodeinfo"
	"github.com/saveio/themis/http/restful"
	"github.com/saveio/themis/http/websocket"
	"github.com/saveio/themis/p2pserver"
	netreqactor "github.com/saveio/themis/p2pserver/actor/req"
	p2p "github.com/saveio/themis/p2pserver/net/protocol"
	"github.com/saveio/themis/txnpool"
	tc "github.com/saveio/themis/txnpool/common"
	"github.com/saveio/themis/txnpool/proc"
	"github.com/saveio/themis/validator/stateful"
	"github.com/saveio/themis/validator/stateless"
	"github.com/urfave/cli"
)

func StartThemis(ctx *cli.Context) {
	initLog(ctx)

	log.Infof("themis version %s", config.Version)

	setMaxOpenFiles()

	cfg, err := InitConfig(ctx)
	if err != nil {
		log.Errorf("initConfig error: %s", err)
		return
	}
	acc, err := InitAccount(ctx)
	if err != nil {
		log.Errorf("initWallet error: %s", err)
		return
	}
	stateHashHeight := config.GetStateHashCheckHeight(cfg.P2PNode.NetworkId)
	ldg, err := InitLedger(ctx, stateHashHeight)
	if err != nil {
		log.Errorf("%s", err)
		return
	}
	txpool, err := InitTxPool(ctx)
	if err != nil {
		log.Errorf("initTxPool error: %s", err)
		return
	}
	p2pSvr, p2p, err := InitP2PNode(ctx, txpool, acc)
	if err != nil {
		log.Errorf("initP2PNode error:%s", err)
		return
	}

	_, err = InitConsensus(ctx, p2p, txpool, acc)
	if err != nil {
		log.Errorf("initConsensus error: %s", err)
		return
	}

	err = InitRpc(ctx)
	if err != nil {
		log.Errorf("initRpc error: %s", err)
		return
	}
	err = InitLocalRpc(ctx)
	if err != nil {
		log.Errorf("initLocalRpc error: %s", err)
		return
	}
	// initGraphQL(ctx)
	InitRestful(ctx)
	InitWs(ctx)
	InitNodeInfo(ctx, p2pSvr)

	go LogCurrBlockHeight()
	waitToExit(ldg)
}

func initLog(ctx *cli.Context) {
	//init log module
	logLevel := ctx.GlobalInt(utils.GetFlagName(utils.LogLevelFlag))
	//if true, the log will not be output to the file
	disableLogFile := ctx.GlobalBool(utils.GetFlagName(utils.DisableLogFileFlag))
	if disableLogFile {
		log.InitLog(logLevel, log.Stdout)
	} else {
		logFileDir := ctx.GlobalString(utils.GetFlagName(utils.LogDirFlag))
		logFileDir = filepath.Join(logFileDir, "") + string(os.PathSeparator)
		alog.InitLog(logFileDir)
		log.InitLog(logLevel, logFileDir, log.Stdout)
	}
}

func InitConfig(ctx *cli.Context) (*config.ThemisConfig, error) {
	//init themis config from cli
	cfg, err := cmd.SetThemisConfig(ctx)
	if err != nil {
		return nil, err
	}
	log.Infof("Config init success")
	return cfg, nil
}

func InitAccount(ctx *cli.Context) (*account.Account, error) {
	if !config.DefConfig.Consensus.EnableConsensus {
		return nil, nil
	}
	walletFile := ctx.GlobalString(utils.GetFlagName(utils.WalletFileFlag))
	if walletFile == "" {
		return nil, fmt.Errorf("please config wallet file using --wallet flag")
	}
	if !common.FileExisted(walletFile) {
		return nil, fmt.Errorf("cannot find wallet file: %s. Please create a wallet first", walletFile)
	}

	acc, err := cmdcom.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("get account error: %s", err)
	}
	pubKey := hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey))
	log.Infof("Using account: %s, pubkey: %s", acc.Address.ToBase58(), pubKey)

	if config.DefConfig.Genesis.ConsensusType == config.CONSENSUS_TYPE_SOLO {
		config.DefConfig.Genesis.SOLO.Bookkeepers = []string{pubKey}
	}

	log.Infof("Account init success")
	return acc, nil
}

func InitLedger(ctx *cli.Context, stateHashHeight uint32) (*ledger.Ledger, error) {
	events.Init() //Init event hub

	var err error
	dbDir := utils.GetStoreDirPath(config.DefConfig.Common.DataDir, config.DefConfig.P2PNode.NetworkName)
	ledger.DefLedger, err = ledger.NewLedger(dbDir, stateHashHeight)
	if err != nil {
		return nil, fmt.Errorf("NewLedger error: %s", err)
	}
	bookKeepers, err := config.DefConfig.GetBookkeepers()
	if err != nil {
		return nil, fmt.Errorf("GetBookkeepers error: %s", err)
	}
	genesisConfig := config.DefConfig.Genesis
	genesisBlock, err := genesis.BuildGenesisBlock(bookKeepers, genesisConfig)
	if err != nil {
		return nil, fmt.Errorf("genesisBlock error %s", err)
	}
	err = ledger.DefLedger.Init(bookKeepers, genesisBlock)
	if err != nil {
		return nil, fmt.Errorf("init ledger error: %s", err)
	}

	log.Infof("Ledger init success")
	return ledger.DefLedger, nil
}

func InitTxPool(ctx *cli.Context) (*proc.TXPoolServer, error) {
	disablePreExec := ctx.GlobalBool(utils.GetFlagName(utils.TxpoolPreExecDisableFlag))
	bactor.DisableSyncVerifyTx = ctx.GlobalBool(utils.GetFlagName(utils.DisableSyncVerifyTxFlag))
	disableBroadcastNetTx := ctx.GlobalBool(utils.GetFlagName(utils.DisableBroadcastNetTxFlag))
	txPoolServer, err := txnpool.StartTxnPoolServer(disablePreExec, disableBroadcastNetTx)
	if err != nil {
		return nil, fmt.Errorf("init txpool error: %s", err)
	}
	stlValidator, _ := stateless.NewValidator("stateless_validator")
	stlValidator.Register(txPoolServer.GetPID(tc.VerifyRspActor))
	stlValidator2, _ := stateless.NewValidator("stateless_validator2")
	stlValidator2.Register(txPoolServer.GetPID(tc.VerifyRspActor))
	stfValidator, _ := stateful.NewValidator("stateful_validator")
	stfValidator.Register(txPoolServer.GetPID(tc.VerifyRspActor))

	bactor.SetTxnPoolPid(txPoolServer.GetPID(tc.TxPoolActor))
	bactor.SetTxPid(txPoolServer.GetPID(tc.TxActor))

	log.Infof("TxPool init success")
	return txPoolServer, nil
}

func InitP2PNode(ctx *cli.Context, txpoolSvr *proc.TXPoolServer, acct *account.Account) (*p2pserver.P2PServer, p2p.P2P, error) {
	if config.DefConfig.Genesis.ConsensusType == config.CONSENSUS_TYPE_SOLO {
		return nil, nil, nil
	}
	p2p, err := p2pserver.NewServer(acct)
	if err != nil {
		return nil, nil, err
	}

	err = p2p.Start()
	if err != nil {
		return nil, nil, fmt.Errorf("p2p service start error %s", err)
	}
	netreqactor.SetTxnPoolPid(txpoolSvr.GetPID(tc.TxActor))
	txpoolSvr.Net = p2p.GetNetwork()
	bactor.SetNetServer(p2p.GetNetwork())
	p2p.WaitForPeersStart()
	log.Infof("P2P init success")
	return p2p, p2p.GetNetwork(), nil
}

func InitConsensus(ctx *cli.Context, p2p p2p.P2P, txpoolSvr *proc.TXPoolServer, acc *account.Account) (consensus.ConsensusService, error) {
	if !config.DefConfig.Consensus.EnableConsensus {
		return nil, nil
	}
	pool := txpoolSvr.GetPID(tc.TxPoolActor)

	consensusType := strings.ToLower(config.DefConfig.Genesis.ConsensusType)
	consensusService, err := consensus.NewConsensusService(consensusType, acc, pool, nil, p2p)
	if err != nil {
		return nil, fmt.Errorf("NewConsensusService %s error: %s", consensusType, err)
	}
	consensusService.Start()

	netreqactor.SetConsensusPid(consensusService.GetPID())
	bactor.SetConsensusPid(consensusService.GetPID())

	log.Infof("Consensus init success")
	return consensusService, nil
}

func InitRpc(ctx *cli.Context) error {
	if !config.DefConfig.Rpc.EnableHttpJsonRpc {
		return nil
	}
	var err error
	exitCh := make(chan interface{}, 0)
	go func() {
		err = jsonrpc.StartRPCServer()
		close(exitCh)
	}()

	flag := false
	select {
	case <-exitCh:
		if !flag {
			return err
		}
	case <-time.After(time.Millisecond * 5):
		flag = true
	}
	log.Infof("Rpc init success")
	return nil
}

func InitLocalRpc(ctx *cli.Context) error {
	if !ctx.GlobalBool(utils.GetFlagName(utils.RPCLocalEnableFlag)) {
		return nil
	}
	var err error
	exitCh := make(chan interface{}, 0)
	go func() {
		err = localrpc.StartLocalServer()
		close(exitCh)
	}()

	flag := false
	select {
	case <-exitCh:
		if !flag {
			return err
		}
	case <-time.After(time.Millisecond * 5):
		flag = true
	}

	log.Infof("Local rpc init success")
	return nil
}

func initGraphQL(ctx *cli.Context) {
	if !config.DefConfig.GraphQL.EnableGraphQL {
		return
	}
	go graphql.StartServer(config.DefConfig.GraphQL)

	log.Infof("GraphQL init success")
}

func InitRestful(ctx *cli.Context) {
	if !config.DefConfig.Restful.EnableHttpRestful {
		return
	}
	go restful.StartServer()

	log.Infof("Restful init success")
}

func InitWs(ctx *cli.Context) {
	if !config.DefConfig.Ws.EnableHttpWs {
		return
	}
	websocket.StartServer()

	log.Infof("Ws init success")
}

func InitNodeInfo(ctx *cli.Context, p2pSvr *p2pserver.P2PServer) {
	// testmode has no p2pserver(see function initP2PNode for detail), simply ignore httpInfoPort in testmode
	if ctx.Bool(utils.GetFlagName(utils.EnableTestModeFlag)) || config.DefConfig.P2PNode.HttpInfoPort == 0 {
		return
	}
	go nodeinfo.StartServer(p2pSvr.GetNetwork())

	log.Infof("Nodeinfo init success")
}

func LogCurrBlockHeight() {
	ticker := time.NewTicker(config.DEFAULT_GEN_BLOCK_TIME * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log.Infof("CurrentBlockHeight = %d", ledger.DefLedger.GetCurrentBlockHeight())
			log.CheckRotateLogFile()
		}
	}
}

func setMaxOpenFiles() {
	max, err := fdlimit.Maximum()
	if err != nil {
		log.Errorf("failed to get maximum open files: %v", err)
		return
	}
	_, err = fdlimit.Raise(uint64(max))
	if err != nil {
		log.Errorf("failed to set maximum open files: %v", err)
		return
	}
}

func waitToExit(db *ledger.Ledger) {
	exit := make(chan bool, 0)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for sig := range sc {
			log.Infof("Themis received exit signal: %v.", sig.String())
			log.Infof("closing ledger...")
			db.Close()
			close(exit)
			break
		}
	}()
	<-exit
}
