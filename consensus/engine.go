package consensus

import (
	"errors"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wire"
)

var AvailableConsensus map[string]ConsensusInterface

type Engine struct {
	sync.Mutex
	cQuit              chan struct{}
	started            bool
	Node               nodeInterface
	ChainConsensusList map[string]ConsensusInterface
	// Chains               map[string]ChainInterface
	CurrentMiningChain   string
	Blockchain           *blockchain.BlockChain
	BlockGen             *blockchain.BlockGenerator
	userMiningPublicKeys map[string]string
	chainCommitteeChange chan string
	// MiningKeys         map[string]string

}

func New(node nodeInterface, blockchain *blockchain.BlockChain, blockgen *blockchain.BlockGenerator) *Engine {
	engine := Engine{
		Node:       node,
		Blockchain: blockchain,
		BlockGen:   blockgen,
	}
	err := engine.LoadMiningKeys(node.GetMiningKeys())
	if err != nil {
		panic(err)
	}
	return &engine
}

//watchConsensusState will watch MiningKey Role as well as chain consensus type
func (engine *Engine) watchConsensusCommittee() {
	Logger.log.Info("start watching consensus committee...")
	engine.chainCommitteeChange = make(chan string)

	for {
		select {
		case <-engine.cQuit:
		case chainName := <-engine.chainCommitteeChange:
			consensusType := engine.Blockchain.Chains[chainName].GetConsensusType()
			userPublicKey, ok := engine.userMiningPublicKeys[consensusType]
			if !ok {
				continue
			}
			if engine.Blockchain.Chains[chainName].GetPubKeyCommitteeIndex(userPublicKey) > 0 {
				if engine.CurrentMiningChain != chainName {
					engine.CurrentMiningChain = chainName
				}
			}
		}
	}
}

func (engine *Engine) Start() error {
	engine.Lock()
	defer engine.Unlock()
	if engine.started {
		return errors.New("Consensus engine is already started")
	}
	Logger.log.Info("starting consensus...")

	engine.cQuit = make(chan struct{})
	go func() {
		go engine.watchConsensusCommittee()
		for {
			select {
			case <-engine.cQuit:
				return
			default:
				time.Sleep(time.Millisecond * 500)
				for chainName, consensus := range engine.ChainConsensusList {
					if chainName == engine.CurrentMiningChain {
						consensus.Start()
					} else {
						consensus.Stop()
					}
				}
			}
		}
	}()
	return nil
}

func (engine *Engine) Stop(name string) error {
	engine.Lock()
	defer engine.Unlock()
	if !engine.started {
		return errors.New("Consensus engine is already stopped")
	}
	engine.started = false
	close(engine.cQuit)
	return nil
}

func (engine *Engine) SwitchConsensus(chainkey string, consensus string) error {
	if engine.ChainConsensusList[common.BEACON_CHAINKEY].GetConsensusName() != engine.Blockchain.BestState.Beacon.ConsensusAlgorithm {
		consensus, ok := AvailableConsensus[engine.ChainConsensusList[common.BEACON_CHAINKEY].GetConsensusName()]
		if ok {
			engine.ChainConsensusList[common.BEACON_CHAINKEY] = consensus.NewInstance()
		} else {
			panic("Update code please")
		}
	}
	for idx := 0; idx < engine.Blockchain.BestState.Beacon.ActiveShards; idx++ {
		shard, ok := engine.Blockchain.BestState.Shard[byte(idx)]
		if ok {
			chainKey := common.GetShardChainKey(byte(idx))
			if shard.ConsensusAlgorithm != engine.ChainConsensusList[chainKey].GetConsensusName() {
				consensus, ok := AvailableConsensus[engine.ChainConsensusList[chainKey].GetConsensusName()]
				if ok {
					engine.ChainConsensusList[chainKey] = consensus.NewInstance()
				} else {
					panic("Update code please")
				}
			}
		} else {
			panic("Oops... Maybe a bug cause this, please update code")
		}
	}
	return nil
}

func RegisterConsensus(name string, consensus ConsensusInterface) error {
	if len(AvailableConsensus) == 0 {
		AvailableConsensus = make(map[string]ConsensusInterface)
	}
	AvailableConsensus[name] = consensus
	return nil
}

func (engine *Engine) ValidateBlockCommitteSig(blockHash *common.Hash, committee []string, validationData string, consensusType string) error {
	return engine.ChainConsensusList[consensusType].ValidateCommitteeSig(blockHash, committee, validationData)
}

func (engine *Engine) IsOngoing(chainName string) bool {
	consensusModule, ok := engine.ChainConsensusList[chainName]
	if ok {
		return consensusModule.IsOngoing()
	}
	return false
}

func (engine *Engine) OnBFTMsg(msg *wire.MessageBFT) {
	if engine.CurrentMiningChain == msg.ChainKey {
		engine.ChainConsensusList[msg.ChainKey].ProcessBFTMsg(msg)
	}
}

func (engine *Engine) GetUserRole() (string, int) {
	if engine.CurrentMiningChain != "" {
		publicKey, _ := engine.GetCurrentMiningPublicKey()
		userRole, _ := engine.Blockchain.Chains[engine.CurrentMiningChain].GetPubkeyRole(publicKey, 0)
		if engine.CurrentMiningChain == common.BEACON_CHAINKEY {
			return userRole, -1
		}
		return userRole, engine.Blockchain.Chains[engine.CurrentMiningChain].GetShardID()
	}
	return "", 0
}

func (engine *Engine) VerifyData(data []byte, sig string, publicKey string, consensusType string) error {
	return nil
}

func (engine *Engine) ValidateProducerSig(block common.BlockInterface, consensusType string) error {
	return nil
}
