package poc

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/saveio/themis/account"
	cutils "github.com/saveio/themis/cmd/utils"
	"github.com/saveio/themis/common/config"
	"github.com/saveio/themis/common/log"
	actorTypes "github.com/saveio/themis/consensus/actor"
	consutils "github.com/saveio/themis/consensus/utils"
	"github.com/saveio/themis/core/ledger"
	"github.com/saveio/themis/core/utils"
	httpcom "github.com/saveio/themis/http/base/common"
	gov "github.com/saveio/themis/smartcontract/service/native/governance"
)

var bufferPool sync.Pool

type State struct {
	sync.Mutex
	View       uint32
	BaseTarget uint64
	Difficulty uint64

	// count how many reader's scoops have been processed
	ProcessedReaderTasks int
	ProcessedNonces      uint64
	ProcessedFakeNonces  uint64

	BestDeadline uint64
	Nonce        uint64
	PlotName     string

	SipVoteInfos     map[uint32]*actorTypes.SipVoteDecision
	ConsVotePubkey   []string
	ConsGovView      uint32
	TriggerConsElect bool

	SubmitView uint32
}

type Miner struct {
	Account      *account.Account
	pocPoolActor *actorTypes.PoCPoolActor
	poolActor    *actorTypes.TxPoolActor
	ledger       *ledger.Ledger
	pid          *actor.PID

	State                   *State
	TotalNonce              uint64
	PlotReader              *Reader
	ReaderTaskCount         int
	RxNonceData             <-chan *NonceData
	AccountId               uint64
	TargetDeadline          uint64
	QueryMiningInfoInterval int
}

func NewMiner(account *account.Account, pocPool, txpool *actor.PID) (*Miner, error) {
	this := &Miner{
		ledger:       ledger.DefLedger,
		Account:      account,
		pocPoolActor: &actorTypes.PoCPoolActor{Pool: pocPool},
		poolActor:    &actorTypes.TxPoolActor{Pool: txpool},
	}

	props := actor.FromProducer(func() actor.Actor {
		return this
	})

	pid, err := actor.SpawnNamed(props, "consensus_poc")
	if err != nil {
		return nil, err
	}
	this.pid = pid

	log.Debug("Miner init successed")

	err = this.Init()
	if err != nil {
		log.Error("Miner Init Error, msg:", err)
		return nil, err
	}

	return this, nil
}

func (self *Miner) Init() error {

	plotDirs := config.DefConfig.PoC.PlotDir
	numWorker := config.DefConfig.PoC.NumWorkTask
	noncesPerCache := config.DefConfig.PoC.NoncesPerCache

	log.Debug("Miner init with numWorker:", numWorker, "noncesPerCache:", noncesPerCache)
	self.AccountId = uint64(cutils.WalletAddressToId([]byte(self.Account.Address.ToBase58())))

	plotStr := strings.TrimSpace(strings.Trim(plotDirs, ","))
	dirs := strings.Split(plotStr, ",")

	PthSep := string(os.PathSeparator)
	totalNumNonce := uint64(0)
	driveIdToPlots := make(map[string]*PlotsDetail)

	for _, plotDir := range dirs {
		plotFiles, err := ioutil.ReadDir(plotDir)
		if err != nil {
			log.Info("Miner init handle dir %s error %s", plotDir, err)
			continue
		}

		for _, plotFile := range plotFiles {
			if plotFile.IsDir() {
				continue
			} else {
				fullPath := plotDir + PthSep + plotFile.Name()
				plot, err := NewPlot(fullPath, false)
				if err != nil {
					log.Info("Miner init handle plot file %s error %s", fullPath, err)
					continue
				}
				if plot.AccountId != self.AccountId {
					log.Info("Miner init skip plot file %s with wrong account id %d", plot.FilePath, plot.AccountId)
					continue
				}

				//skip Device ID currently
				deviceId := plotDir
				if strings.LastIndex(deviceId, PthSep) < len(deviceId)-1 {
					deviceId = deviceId + PthSep
				}

				if _, ok := driveIdToPlots[deviceId]; !ok {
					driveIdToPlots[deviceId] = &PlotsDetail{Lookup: make(map[string]*Plot)}
					log.Debug("Miner init handle dir:", deviceId)
				}

				driveIdToPlots[deviceId].Plots = append(driveIdToPlots[deviceId].Plots, plot)
				driveIdToPlots[deviceId].Lookup[fullPath] = plot
				log.Debug("Miner init handle plot file:", fullPath)

				totalNumNonce += plot.Nonces
			}
		}
	}
	self.TotalNonce = totalNumNonce
	if self.TotalNonce == 0 {
		return fmt.Errorf("Miner init fail to find any plot files!")
	}

	bufferCount := numWorker * 2
	bufferSize := noncesPerCache * utils.SCOOP_SIZE
	chanReadReplies := make(chan *ReadReply, bufferCount)
	chanNonceData := make(chan *NonceData, numWorker)
	chanEmptyBuffers := make(chan struct{})

	self.RxNonceData = chanNonceData
	for i := 0; i < numWorker; i++ {
		go WorkTask(chanReadReplies, chanEmptyBuffers, chanNonceData)
	}
	self.PlotReader, _ = NewReader(driveIdToPlots, chanReadReplies, bufferSize)
	self.ReaderTaskCount = len(driveIdToPlots)
	self.AccountId = uint64(cutils.WalletAddressToId([]byte(self.Account.Address.ToBase58())))
	self.QueryMiningInfoInterval = config.DefConfig.PoC.QueryMiningInfoInterval
	self.TargetDeadline = math.MaxUint64
	self.RxNonceData = chanNonceData

	state := &State{
		View:                 0,
		BestDeadline:         math.MaxUint64,
		BaseTarget:           1,
		ProcessedReaderTasks: 0,
		ProcessedNonces:      0,
		SipVoteInfos:         make(map[uint32]*actorTypes.SipVoteDecision),
	}
	self.State = state

	//init pool
	bufferPool.New = func() interface{} {
		b := make([]byte, bufferSize)
		return b
	}

	return nil
}

func (self *Miner) RemovePlots(plotfile string) error {
	dir, file := filepath.Split(plotfile)
	log.Debugf("Miner RemovePlots dir:%s, file:%s", dir, file)

	if detail, ok := self.PlotReader.DriveIdToPlots[dir]; ok {
		plot := detail.Lookup[plotfile]
		self.State.Lock()
		self.TotalNonce -= plot.Nonces
		self.State.Unlock()
		plot.Removed = true
		log.Debug("Miner RemovePlots :", plotfile)
	}
	return nil
}

func (self *Miner) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Restarting:
		log.Info("poc actor restarting")
	case *actor.Stopping:
		log.Info("poc actor stopping")
	case *actor.Stopped:
		log.Info("poc actor stopped")
	case *actor.Started:
		log.Info("poc actor started")
	case *actor.Restart:
		log.Info("poc actor restart")
	case *actorTypes.StartConsensus:
		log.Info("poc actor start consensus")
	case *actorTypes.StopConsensus:
		self.stop()
	case *actorTypes.PlotFileAction:
		log.Debug("PoC miner receive remove plot info", msg)
		self.RemovePlots(msg.PlotFile)
	case *actorTypes.SipVoteDecision:
		self.State.Lock()
		self.State.SipVoteInfos[msg.SipIndex] = msg
		self.State.Unlock()
		log.Debug("PoC miner receive sip vote info", msg)
	case *actorTypes.ConsVoteDecision:
		self.State.Lock()
		self.State.ConsVotePubkey = msg.NodesPubkey
		govView, err := consutils.GetConsGovView()
		if err != nil {
			log.Debug("PoC miner get cons gov view err", err)
		} else {
			self.State.ConsGovView = govView.GovView
		}
		self.State.Unlock()
		log.Debug("PoC miner receive consensus vote", msg)
	case *actorTypes.TriggerConsElect:
		self.State.Lock()
		self.State.TriggerConsElect = true
		self.State.Unlock()
		log.Debug("PoC miner receive request for moving up consensus elect")
	default:
		log.Info("vbft actor: Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

func (self *Miner) GetPID() *actor.PID {
	return self.pid
}

func (self *Miner) Start() error {
	go self.ScheduleTask()
	go self.SubmitTask()
	return nil
}

func (self *Miner) Halt() error {
	self.pid.Tell(&actorTypes.StopConsensus{})
	return nil
}

func (self *Miner) stop() error {
	return nil
}

func (self *Miner) ScheduleTask() {
	log.Debug("ScheduleTask start, query mining info interval", self.QueryMiningInfoInterval)

	//start mining as soon as possible!
	t1 := time.NewTimer(time.Millisecond * time.Duration(100))

	for {
		select {
		case <-t1.C:
			t1.Reset(time.Second * time.Duration(self.QueryMiningInfoInterval))
			log.Debug("Try get miner information")

			miningInfo, err := consutils.GetMiningInfo()
			if err != nil {
				log.Debug("getMiningInfo err %s", err)
				continue
			}

			self.State.Lock()
			oldView := self.State.View
			self.State.Unlock()

			//detect view change during interval
			if miningInfo.View > oldView {
				log.Debug("Get new miner information")
				log.Debug("View ", miningInfo.View)
				log.Debug("BaseTarget ", miningInfo.BaseTarget)
				log.Debug("GenerationSignature ", miningInfo.GenerationSignature)

				self.State.Lock()
				self.State.View = miningInfo.View

				//don't mining during catchup
				if oldView > 0 && miningInfo.View > oldView+1 {
					self.State.Unlock()
					continue
				}

				self.State.BestDeadline = math.MaxUint64
				self.State.BaseTarget = uint64(miningInfo.BaseTarget)
				self.State.PlotName = ""

				gensig := miningInfo.GenerationSignature[:]
				scoop := utils.CalculateScoop(uint64(self.State.View), gensig)
				log.Debug("calculateScoop return", scoop)

				self.State.ProcessedReaderTasks = 0
				self.State.ProcessedNonces = 0
				self.State.ProcessedFakeNonces = 0

				//remove sip vote out of date
				height := ledger.DefLedger.GetCurrentBlockHeight()
				for index, v := range self.State.SipVoteInfos {
					sip, err := httpcom.GetSipInfo(v.SipIndex)
					if err != nil {
						log.Debug("ScheduleTask get sip %d fail", v.SipIndex)
						continue
					}

					if height > sip.RegHeight+gov.SIP_VOTE_DELAY+gov.SIP_VOTE_PERIOD {
						delete(self.State.SipVoteInfos, index)
					}
				}

				//clear cons vote after consensus gov view change
				govView, err := consutils.GetConsGovView()
				if err != nil {
					log.Debug("PoC miner get cons gov view err", err)
					self.State.Unlock()
					continue
				}
				if self.State.ConsGovView != govView.GovView {
					self.State.ConsVotePubkey = []string{}
				}
				self.State.Unlock()

				self.PlotReader.startReading(self.State.View, scoop, gensig)
			}
		}
	}
}

func (self *Miner) SubmitTask() {
	for {
		select {
		case nonceData := <-self.RxNonceData:
			self.State.Lock()

			//task of last view not finish util new view!
			if self.State.View > nonceData.View {
				self.State.Unlock()
				continue
			}

			//already submit for this view
			if self.State.SubmitView == nonceData.View {
				self.State.Unlock()
				continue
			}

			deadline := nonceData.Deadline
			log.Debug("SubmitTask get deadline", deadline)

			if deadline < self.State.BestDeadline && deadline < self.TargetDeadline {
				self.State.BestDeadline = deadline
				self.State.Nonce = nonceData.Nonce
				self.State.PlotName = nonceData.PlotName
				log.Debugf("SubmitTask get new best deadline %d, nonce %d from plot %s",
					deadline, nonceData.Nonce, nonceData.PlotName)
			}

			if nonceData.ReaderTaskProcessed {
				self.State.ProcessedReaderTasks++
				log.Debugf("SubmitTask number finished %d read task", self.State.ProcessedReaderTasks)
			}

			self.State.ProcessedNonces += nonceData.NumNonce
			if deadline == math.MaxUint64 {
				self.State.ProcessedFakeNonces += nonceData.NumNonce
			}

			log.Debugf("SubmitTask processed %d nonce, %d fake nonce, total nonce %d", self.State.ProcessedNonces, self.State.ProcessedFakeNonces, self.TotalNonce)

			if self.State.ProcessedNonces-self.State.ProcessedFakeNonces >= self.TotalNonce {

				// push result to poc pool
				param := &gov.SubmitNonceParam{
					View:        self.State.View,
					Address:     self.Account.Address,
					Id:          int64(self.AccountId),
					Nonce:       self.State.Nonce,
					Deadline:    self.State.BestDeadline,
					PlotName:    self.State.PlotName,
					MoveUpElect: self.State.TriggerConsElect,
				}

				//add vote info
				param.VoteConsPub = make([]string, len(self.State.ConsVotePubkey))
				copy(param.VoteConsPub, self.State.ConsVotePubkey)

				log.Debug("SubmitTask consvote pubkey", param.VoteConsPub)

				for _, v := range self.State.SipVoteInfos {
					param.VoteId = append(param.VoteId, v.SipIndex)
					param.VoteInfo = append(param.VoteInfo, byte(v.Agree))
				}

				self.pocPoolActor.PushPoCParam(param)
				self.State.SubmitView = nonceData.View
				log.Debug("SubmitTask submit nonce for view ", self.State.View)

				//clear consensus elect flag after sending once
				self.State.TriggerConsElect = false
			}

			self.State.Unlock()
		}
	}
}
