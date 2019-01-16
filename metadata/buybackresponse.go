package metadata

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
)

type BuyBackResponse struct {
	MetadataBase
	RequestedTxID common.Hash
}

func NewBuyBackResponse(requestedTxID common.Hash, metaType int) *BuyBackResponse {
	metadataBase := MetadataBase{
		Type: metaType,
	}
	return &BuyBackResponse{
		RequestedTxID: requestedTxID,
		MetadataBase:  metadataBase,
	}
}

func (bbRes *BuyBackResponse) CheckTransactionFee(tr Transaction, minFee uint64) bool {
	// no need to have fee for this tx
	return common.TrueValue
}

func (bbRes *BuyBackResponse) ValidateTxWithBlockChain(txr Transaction, bcr BlockchainRetriever, chainID byte, db database.DatabaseInterface) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with requeste tx (via RequestedTxID) in current block
	return common.FalseValue, nil
}

func (bbRes *BuyBackResponse) ValidateSanityData(bcr BlockchainRetriever, txr Transaction) (bool, bool, error) {
	return common.FalseValue, common.TrueValue, nil
}

func (bbRes *BuyBackResponse) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning common.TrueValue here
	return common.TrueValue
}

func (bbRes *BuyBackResponse) Hash() *common.Hash {
	record := bbRes.RequestedTxID.String()
	record += bbRes.MetadataBase.Hash().String()
	// final hash
	record += bbRes.MetadataBase.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash
}
