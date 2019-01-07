package rpcserver

import (
	"encoding/json"

	params2 "github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/ninjadotorg/constant/wire"
)

func (self RpcServer) handleGetBondTypes(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	tempRes1 := jsonresult.GetBondTypeResultItem{
		StartSellingAt: 0,
		EndSellingAt:   500,
		Maturity:       700,
		BuyBackPrice:   110, // in constant
		BuyPrice:       105, // in constant
		TotalIssue:     1000,
		Available:      800,
	}
	tempRes2 := jsonresult.GetBondTypeResultItem{
		StartSellingAt: 0,
		EndSellingAt:   500,
		Maturity:       700,
		BuyBackPrice:   130, // in constant
		BuyPrice:       110, // in constant
		TotalIssue:     2000,
		Available:      800,
	}
	result := jsonresult.GetBondTypeResult{
		BondTypes: make(map[string]jsonresult.GetBondTypeResultItem),
	}

	result.BondTypes["fc8bbbd183f97ff6cc55a62b2ddceade8e93eed5cdf1240b42e223d589b29645"] = tempRes1

	result.BondTypes["fe7d3d124cf0309d8f1575982842b57266951a37a717a4d332a69eb176c409fa"] = tempRes2

	return result, nil
}

func (self RpcServer) handleGetGOVParams(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	govParam := self.config.BlockChain.BestState[0].BestBlock.Header.GOVConstitution.GOVParams
	return govParam, nil
}

func (self RpcServer) handleGetGOVConstitution(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constitution := self.config.BlockChain.BestState[0].BestBlock.Header.GOVConstitution
	return constitution, nil
}

func (self RpcServer) handleGetListGOVBoard(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return self.config.BlockChain.BestState[0].BestBlock.Header.GOVGovernor.BoardPubKeys, nil
}

func (self RpcServer) handleCreateRawTxWithBuyBackRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	arrayParams := common.InterfaceSlice(params)

	// Req param #4: issuing request info
	buyBackReq := arrayParams[4].(map[string]interface{})
	voutIndex := int(buyBackReq["voutIndex"].(float64))
	buyBackFromTxIDBytes := []byte(buyBackReq["buyBackFromTxID"].(string))
	buyBackFromTxID := common.Hash{}
	copy(buyBackFromTxID[:], buyBackFromTxIDBytes)
	metaType := metadata.BuyBackRequestMeta
	meta := metadata.NewBuyBackRequest(
		buyBackFromTxID,
		voutIndex,
		metaType,
	)
	normalTx, err := self.buildRawTransaction(params, meta)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err := json.Marshal(normalTx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            normalTx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (self RpcServer) handleCreateAndSendTxWithBuyBackRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawTxWithBuyBackRequest(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := self.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID: sendResult.(jsonresult.CreateTransactionResult).TxID,
	}
	return result, nil
}

func (self RpcServer) handleCreateRawTxWithBuySellRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	arrayParams := common.InterfaceSlice(params)

	// Req param #4: buy/sell request info
	buySellReq := arrayParams[4].(map[string]interface{})

	paymentAddressMap := buySellReq["paymentAddress"].(map[string]interface{})
	paymentAddress := privacy.PaymentAddress{
		Pk: []byte(paymentAddressMap["pk"].(string)),
		Tk: []byte(paymentAddressMap["tk"].(string)),
	}
	assetTypeBytes := []byte(buySellReq["assetType"].(string))
	assetType := common.Hash{}
	copy(assetType[:], assetTypeBytes)
	amount := uint64(buySellReq["amount"].(float64))
	buyPrice := uint64(buySellReq["buyPrice"].(float64))
	metaType := metadata.BuyFromGOVRequestMeta
	meta := metadata.NewBuySellRequest(
		paymentAddress,
		assetType,
		amount,
		buyPrice,
		metaType,
	)
	normalTx, err := self.buildRawTransaction(params, meta)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err := json.Marshal(normalTx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            normalTx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (self RpcServer) handleCreateAndSendTxWithBuySellRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawTxWithBuySellRequest(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := self.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID: sendResult.(jsonresult.CreateTransactionResult).TxID,
	}
	return result, nil
}

func (self RpcServer) buildRawSealLv3VoteGOVProposalTransaction(
	params interface{},
) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	voteInfo := arrayParams[len(arrayParams)-4]
	firstPubKey := arrayParams[len(arrayParams)-3] // firstPubKey is pubkey of itself
	secondPubKey := arrayParams[len(arrayParams)-2]
	thirdPubKey := arrayParams[len(arrayParams)-1]
	Seal3Data := common.Encrypt(common.Encrypt(common.Encrypt(voteInfo, thirdPubKey), secondPubKey), firstPubKey)
	meta := metadata.NewSealedLv3GOVBallotMetadata([]byte(Seal3Data.(string)), [][]byte{[]byte(firstPubKey.(string)), []byte(secondPubKey.(string)), []byte(thirdPubKey.(string))})
	tx, err := self.buildRawTransaction(params, meta)
	return tx, err
}

func (self RpcServer) handleCreateRawSealLv3VoteGOVProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	tx, err := self.buildRawSealLv3VoteGOVProposalTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err := json.Marshal(tx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

//create lv3 vote by 3 layer encrypt
func (self RpcServer) handleCreateAndSendSealLv3VoteGOVProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawSealLv3VoteGOVProposalTransaction(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txId, err := self.handleSendRawTransaction(newParam, closeChan)
	return txId, err
}

func (self RpcServer) buildRawSealLv2VoteGOVProposalTransaction(
	params interface{},
) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	PointerToLv3Ballot := arrayParams[len(arrayParams)-6]
	Seal3Data := arrayParams[len(arrayParams)-5]
	firstPubKey := arrayParams[len(arrayParams)-4]
	secondPubKey := arrayParams[len(arrayParams)-3]
	thirdPubKey := arrayParams[len(arrayParams)-2]
	firstPrivateKey := arrayParams[len(arrayParams)-1]
	Seal2Data := common.Decrypt(Seal3Data, firstPrivateKey)
	Pointer := common.Hash{}
	copy(Pointer[:], []byte(PointerToLv3Ballot.(string)))
	meta := metadata.NewSealedLv2GOVBallotMetadata([]byte(Seal2Data.(string)), [][]byte{[]byte(firstPubKey.(string)), []byte(secondPubKey.(string)), []byte(thirdPubKey.(string))}, Pointer)
	tx, err := self.buildRawTransaction(params, meta)
	return tx, err
}

func (self RpcServer) handleCreateRawSealLv2VoteGOVProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	tx, err := self.buildRawSealLv2VoteGOVProposalTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err := json.Marshal(tx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

//create lv2 vote by decrypt A layer
func (self RpcServer) handleCreateAndSendSealLv2VoteGOVProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawSealLv2VoteGOVProposalTransaction(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txId, err := self.handleSendRawTransaction(newParam, closeChan)
	return txId, err
}

func (self RpcServer) buildRawSealLv1VoteGOVProposalTransaction(
	params interface{},
) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)

	PointerToLv3Ballot := arrayParams[len(arrayParams)-7]
	Pointer3 := common.Hash{}
	copy(Pointer3[:], []byte(PointerToLv3Ballot.(string)))

	PointerToLv2Ballot := arrayParams[len(arrayParams)-6]
	Pointer2 := common.Hash{}
	copy(Pointer2[:], []byte(PointerToLv2Ballot.(string)))

	Seal2Data := arrayParams[len(arrayParams)-5]
	firstPubKey := arrayParams[len(arrayParams)-4]
	secondPubKey := arrayParams[len(arrayParams)-3]
	thirdPubKey := arrayParams[len(arrayParams)-2]
	secondPrivateKey := arrayParams[len(arrayParams)-1]
	Seal1Data := common.Decrypt(Seal2Data, secondPrivateKey)
	meta := metadata.NewSealedLv1GOVBallotMetadata([]byte(Seal1Data.(string)), [][]byte{[]byte(firstPubKey.(string)), []byte(secondPubKey.(string)), []byte(thirdPubKey.(string))}, Pointer2, Pointer3)
	tx, err := self.buildRawTransaction(params, meta)
	return tx, err
}

func (self RpcServer) handleCreateRawSealLv1VoteGOVProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	tx, err := self.buildRawSealLv1VoteGOVProposalTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err := json.Marshal(tx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (self RpcServer) handleCreateAndSendSealLv1VoteGOVProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawSealLv1VoteGOVProposalTransaction(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txId, err := self.handleSendRawTransaction(newParam, closeChan)
	return txId, err
}

func (self RpcServer) buildRawNormalVoteGOVProposalTransactionFromSealer(
	params interface{},
) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)

	PointerToLv3Ballot := arrayParams[len(arrayParams)-7]
	Pointer3 := common.Hash{}
	copy(Pointer3[:], []byte(PointerToLv3Ballot.(string)))

	PointerToLv1Ballot := arrayParams[len(arrayParams)-6]
	Pointer1 := common.Hash{}
	copy(Pointer1[:], []byte(PointerToLv1Ballot.(string)))

	Seal1Data := arrayParams[len(arrayParams)-5]
	firstPubKey := arrayParams[len(arrayParams)-4]
	secondPubKey := arrayParams[len(arrayParams)-3]
	thirdPubKey := arrayParams[len(arrayParams)-2]
	thirdPrivateKey := arrayParams[len(arrayParams)-1]
	normalBallot := common.Decrypt(Seal1Data, thirdPrivateKey)
	meta := metadata.NewNormalGOVBallotFromSealerMetadata(normalBallot.([]byte), [][]byte{firstPubKey.([]byte), secondPubKey.([]byte), thirdPubKey.([]byte)}, Pointer1, Pointer3)
	tx, err := self.buildRawTransaction(params, meta)
	return tx, err
}

func (self RpcServer) handleCreateRawNormalVoteGOVProposalTransactionFromSealer(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	tx, err := self.buildRawNormalVoteGOVProposalTransactionFromSealer(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err := json.Marshal(tx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (self RpcServer) handleCreateAndSendNormalVoteGOVProposalTransactionFromSealer(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawNormalVoteGOVProposalTransactionFromSealer(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txId, err := self.handleSendRawTransaction(newParam, closeChan)
	return txId, err
}

func (self RpcServer) buildRawNormalVoteGOVProposalTransactionFromOwner(
	params interface{},
) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	PointerToLv3Ballot := arrayParams[len(arrayParams)-5]
	Pointer := common.Hash{}
	copy(Pointer[:], []byte(PointerToLv3Ballot.(string)))
	firstPubKey := arrayParams[len(arrayParams)-4]
	secondPubKey := arrayParams[len(arrayParams)-3]
	thirdPubKey := arrayParams[len(arrayParams)-2]
	normalBallot := arrayParams[len(arrayParams)-1]
	meta := metadata.NewNormalGOVBallotFromOwnerMetadata(normalBallot.([]byte), [][]byte{[]byte(firstPubKey.(string)), []byte(secondPubKey.(string)), []byte(thirdPubKey.(string))}, Pointer)
	tx, err := self.buildRawTransaction(params, meta)
	return tx, err
}

func (self RpcServer) handleCreateRawNormalVoteGOVProposalTransactionFromOwner(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	tx, err := self.buildRawNormalVoteGOVProposalTransactionFromOwner(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err := json.Marshal(tx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (self RpcServer) handleCreateAndSendNormalVoteGOVProposalTransactionFromOwner(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawNormalVoteGOVProposalTransactionFromOwner(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txId, err := self.handleSendRawTransaction(newParam, closeChan)
	return txId, err
}

func (self RpcServer) buildRawVoteGOVBoardTransaction(
	params interface{},
) (*transaction.TxCustomToken, error) {
	arrayParams := common.InterfaceSlice(params)
	candidatePaymentAddress := arrayParams[len(arrayParams)-1].(string)
	account, _ := wallet.Base58CheckDeserialize(candidatePaymentAddress)
	metadata := metadata.NewVoteGOVBoardMetadata(account.KeySet.PaymentAddress.Pk)
	tx, err := self.buildRawCustomTokenTransaction(params, metadata)
	return tx, err
}

func (self RpcServer) handleSendRawVoteBoardGOVTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	base58CheckData := arrayParams[0].(string)
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckData)

	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	tx := transaction.TxCustomToken{}
	// Logger.log.Info(string(rawTxBytes))
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	hash, txDesc, err := self.config.TxMemPool.MaybeAcceptTransaction(&tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	Logger.log.Infof("there is hash of transaction: %s\n", hash.String())
	Logger.log.Infof("there is priority of transaction in pool: %d", txDesc.StartingPriority)

	// broadcast message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdCustomToken)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	self.config.Server.PushMessageToAll(txMsg)

	return tx.Hash(), nil
}

func (self RpcServer) handleCreateRawVoteGOVBoardTransaction(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *RPCError) {
	var err error
	tx, err := self.buildRawVoteGOVBoardTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err := json.Marshal(tx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		//HexData:         hexData,
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (self RpcServer) handleCreateAndSendVoteGOVBoardTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawVoteGOVBoardTransaction(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txId, err := self.handleSendRawVoteBoardGOVTransaction(newParam, closeChan)
	return txId, err
}

func (self RpcServer) buildRawSubmitGOVProposalTransaction(
	params interface{},
) (*transaction.Tx, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	NParams := len(arrayParams)
	GOVParams := *params2.NewGOVParamsFromRPC(arrayParams[NParams-4])
	executeDuration := arrayParams[NParams-3].(uint32)
	explanation := arrayParams[NParams-2].(string)
	paymentData := common.InterfaceSlice(arrayParams[NParams-1])
	address := privacy.PaymentAddress{
		Pk: paymentData[0].([]byte),
		Tk: paymentData[1].([]byte),
	}

	meta := metadata.NewSubmitGOVProposalMetadata(GOVParams, executeDuration, explanation, &address)
	tx, err := self.buildRawTransaction(params, meta)
	return tx, err
}

func (self RpcServer) handleCreateRawSubmitGOVProposalTransaction(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *RPCError) {
	var err error
	tx, err := self.buildRawSubmitGOVProposalTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err := json.Marshal(tx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		//HexData:         hexData,
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (self RpcServer) handleSendRawSubmitGOVProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Info(params)
	var err error
	arrayParams := common.InterfaceSlice(params)
	base58CheckData := arrayParams[0].(string)
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckData)

	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	tx := transaction.Tx{}
	// Logger.log.Info(string(rawTxBytes))
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	hash, txDesc, err := self.config.TxMemPool.MaybeAcceptTransaction(&tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	Logger.log.Infof("there is hash of transaction: %s\n", hash.String())
	Logger.log.Infof("there is priority of transaction in pool: %d", txDesc.StartingPriority)

	// broadcast message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdTx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	self.config.Server.PushMessageToAll(txMsg)

	return tx.Hash(), nil
}

func (self RpcServer) handleCreateAndSendSubmitGOVProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawSubmitGOVProposalTransaction(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txId, err := self.handleSendRawSubmitGOVProposalTransaction(newParam, closeChan)
	return txId, err
}

func (self RpcServer) handleCreateRawTxWithOracleFeed(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	arrayParams := common.InterfaceSlice(params)
	// Req param #4: oracle feed
	oracleFeed := arrayParams[4].(map[string]interface{})

	assetTypeBytes := []byte(oracleFeed["assetType"].(string))
	assetType := common.Hash{}
	copy(assetType[:], assetTypeBytes)
	price := uint64(oracleFeed["price"].(float64))
	metaType := metadata.OracleFeedMeta
	meta := metadata.NewOracleFeed(
		assetType,
		price,
		metaType,
	)

	normalTx, err := self.buildRawTransaction(params, meta)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err := json.Marshal(normalTx)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            normalTx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (self RpcServer) handleCreateAndSendTxWithOracleFeed(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawTxWithOracleFeed(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := self.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID: sendResult.(jsonresult.CreateTransactionResult).TxID,
	}
	return result, nil
}

func (self RpcServer) handleCreateRawTxWithUpdatingOracleBoard(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var err error
	arrayParams := common.InterfaceSlice(params)
	// Req param #4: updating oracle board info
	updatingOracleBoard := arrayParams[4].(map[string]interface{})
	action := int8(updatingOracleBoard["action"].(float64))
	oraclePubKeys := updatingOracleBoard["oraclePubKeys"].([]interface{})
	assertedOraclePKs := [][]byte{}
	for _, pk := range oraclePubKeys {
		assertedOraclePKs = append(assertedOraclePKs, []byte(pk.(string)))
	}
	signs := updatingOracleBoard["signs"].(map[string]interface{})
	assertedSigns := map[string][]byte{}
	for k, s := range signs {
		assertedSigns[k] = []byte(s.(string))
	}
	metaType := metadata.UpdatingOracleBoardMeta
	meta := metadata.NewUpdatingOracleBoard(
		action,
		assertedOraclePKs,
		assertedSigns,
		metaType,
	)

	normalTx, err := self.buildRawTransaction(params, meta)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	byteArrays, err := json.Marshal(normalTx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            normalTx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (self RpcServer) handleCreateAndSendTxWithUpdatingOracleBoard(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawTxWithUpdatingOracleBoard(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := self.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	result := jsonresult.CreateTransactionResult{
		TxID: sendResult.(jsonresult.CreateTransactionResult).TxID,
	}
	return result, nil
}
