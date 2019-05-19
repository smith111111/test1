package wallet

import (
	"context"
	"sync"

	"gcwallet/eth-wallet-interface"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
	"gcwallet/go-ethwallet/etherscan"
	"github.com/ethereum/go-ethereum/common"
	"github.com/op/go-logging"
	"time"
	"strconv"
)

var Log = logging.MustGetLogger("WalletService")

// Service - used to represent WalletService
type Service struct {
	db       wallet.Datastore
	client   *ethclient.Client
	coinType wallet.CoinType

	chainHeight uint32
	bestBlock   string

	lock sync.RWMutex

	doneChan chan struct{}

	// eherscan api
	addr 		  common.Address
	etherscan 	  *etherscan.API
	contract 		  string
}

const nullHash = "0000000000000000000000000000000000000000000000000000000000000000"

// NewWalletService - used to create new wallet service
func NewWalletService(db wallet.Datastore, client *ethclient.Client, coinType wallet.CoinType, addr common.Address, ea *etherscan.API, contract string) *Service {
	return &Service{db, client, coinType, 0, nullHash, sync.RWMutex{}, make(chan struct{}), addr, ea, contract}
}

// Start - the wallet daemon
func (ws *Service) Start() {
	log.Infof("Starting %s WalletService", ws.coinType.String())
	go ws.UpdateState()
}

// Stop - the wallet daemon
func (ws *Service) Stop() {
	ws.doneChan <- struct{}{}
}

// ChainTip - get the chain tip
func (ws *Service) ChainTip() (uint32, chainhash.Hash) {
	ws.lock.RLock()
	defer ws.lock.RUnlock()
	ch, _ := chainhash.NewHashFromStr(ws.bestBlock)
	return uint32(ws.chainHeight), *ch
}

// UpdateState - updates state
func (ws *Service) UpdateState() {
	// Start by fetching the chain height from the API
	log.Debugf("querying for %s chain height", ws.coinType.String())
	best, err := ws.client.HeaderByNumber(context.Background(), nil)
	if err == nil {
		log.Debugf("%s chain height: %d", ws.coinType.String(), best.Nonce)
		ws.lock.Lock()
		ws.chainHeight = uint32(best.Number.Uint64())
		ws.bestBlock = best.TxHash.String()
		ws.lock.Unlock()
	} else {
		log.Errorf("error querying API for chain height: %s", err.Error())
	}

	go ws.syncTxs([]common.Address{ws.addr})
}

func (ws *Service) GetNomalTransactions(addrs []string) ([]*Transaction, error) {
	var transactions []*Transaction
	txs, err := ws.etherscan.GetNormalTransactions(addrs...)
	if err != nil {
		return transactions, err
	}

	for _, tx := range txs {
		//blockNumber, err := strconv.ParseInt(strings.Trim(tx.BlockNumber, `"`), 10, 64)
		//if err != nil {
		//	return transactions, err
		//}
		transactions = append(transactions, &Transaction{
			//BlockNumber: int32(blockNumber),
			TimeStamp: tx.TimeStamp.Time,
			Hash: tx.Hash.Hash,
			Nonce: tx.Nonce,
			BlockHash: tx.BlockHash.Hash,
			From: tx.From.Address,
			To: tx.To.Address,
			Value: tx.Value,
			Gas: tx.Gas,
			GasPrice: tx.GasPrice,
			GasUsed: tx.GasUsed,
			TxreceiptStatus: tx.TxreceiptStatus.String(),
			TransactionIndex: tx.TransactionIndex,
			Input: tx.Input,
			ContractAddress: tx.ContractAddress.Address,
			CumulativeGasUsed: tx.CumulativeGasUsed,
			Confirmations: tx.Confirmations,
		})
	}

	return transactions, nil
}

// 获取ERC20代币交易记录
func (ws *Service) GetERC20TokenTransactions(addr string) ([]*Transaction, error) {
	var transactions []*Transaction
	txs, err := ws.etherscan.GetERC20TokenTransactions(addr, ws.contract)
	if err != nil {
		return transactions, err
	}

	for _, tx := range txs {
		//blockNumber, err := strconv.ParseInt(strings.Trim(tx.BlockNumber, `"`), 10, 64)
		//if err != nil {
		//	return transactions, err
		//}
		transactions = append(transactions, &Transaction{
			//BlockNumber: int32(blockNumber),
			TimeStamp: tx.TimeStamp.Time,
			Hash: tx.Hash.Hash,
			Nonce: tx.Nonce,
			BlockHash: tx.BlockHash.Hash,
			From: tx.From.Address,
			To: tx.To.Address,
			Value: tx.Value,
			Gas: tx.Gas,
			GasPrice: tx.GasPrice,
			GasUsed: tx.GasUsed,
			TxreceiptStatus: "success",
			TransactionIndex: tx.TransactionIndex,
			Input: tx.Input,
			ContractAddress: tx.ContractAddress.Address,
			CumulativeGasUsed: tx.CumulativeGasUsed,
			Confirmations: tx.Confirmations,
		})
	}

	return transactions, nil
}

// Query API for TXs and synchronize db state
func (ws *Service) syncTxs(addrs []common.Address) {
	Log.Debugf("Querying for %s utxos", ws.coinType.String())
	var query []string
	for _, addr := range addrs {
		query = append(query, addr.String())
	}

	var (
		txs []*Transaction
		err error
	)

	if ws.contract != "" {
		txs, err = ws.GetERC20TokenTransactions(query[0])
	} else {
		txs, err = ws.GetNomalTransactions(query)
	}

	if err != nil {
		Log.Errorf("Error downloading txs for %s: %s", ws.coinType.String(), err.Error())
	} else {
		Log.Debugf("Downloaded %d %s transactions", len(txs), ws.coinType.String())
		ws.saveTxsToDB(txs)
	}
}

// For each API response we will need to determine the net coins leaving/entering the wallet as well as determine
// if the transaction was exclusively for our `watch only` addresses. We will also build a Tx object suitable
// for saving to the db and delete any existing txs not returned by the API. Finally, for any output matching a key
// in our wallet we need to mark that key as used in the db
func (ws *Service) saveTxsToDB(txns []*Transaction) {
	ws.lock.RLock()
	chainHeight := int32(ws.chainHeight)
	ws.lock.RUnlock()

	// Iterate over new txs and put them to the db
	for _, t := range txns {
		ws.saveSingleTxToDB(t, chainHeight)
	}
}

func (ws *Service) saveSingleTxToDB(t *Transaction, chainHeight int32) {
	height := int32(0)

	confirmations, err := strconv.ParseInt(t.Confirmations, 10, 64)
	if err != nil {
		Log.Errorf("error converting to parse int for %s: %s", ws.coinType.String(), err.Error())
	}
	if confirmations > 0 {
		height = chainHeight - (int32(confirmations) - 1)
	}

	if _, err := ws.db.Txns().Get(t.Hash.String());
	err != nil {
		ts := time.Now()
		if confirmations > 0 {
			ts = t.TimeStamp
		}

		ws.db.Txns().Put(t.Hash.String(), t.Value, t.From.String(), t.To.String(), t.GasUsed, t.TxreceiptStatus, t.ContractAddress.String(), height, ts.Unix(), t.Input)
	} else {
		ws.db.Txns().UpdateHeight(t.Hash.String(), height, t.TimeStamp.Unix())
	}
}
