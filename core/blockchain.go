package core

import (
  "encoding/hex"
  "github.com/boltdb/bolt"
)

const dbFile = "chain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "Genesis block made by Saehan"

type Blockchain struct {
  tip []byte
  db  *bolt.DB
}

type BlockchainIterator struct {
  currentHash []byte
  db          *bolt.DB
}

func NewGenesisBlock(coinbase *Transaction) *Block {
  return NewBlock([]*Transaction{coinbase}, []byte{})
}

func NewBlockchain(address string) *Blockchain {
  var tip []byte
  db, _ := bolt.Open(dbFile, 0600, nil)

  db.Update(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte(blocksBucket))

    if b == nil {
      cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
      genesis := NewGenesisBlock(cbtx)
      b, _ := tx.CreateBucket([]byte(blocksBucket))
      b.Put(genesis.Hash, genesis.Serialize())
      b.Put([]byte("l"), genesis.Hash)
      tip = genesis.Hash
    } else {
      tip = b.Get([]byte("l"))
    }

    return nil
  })

  bc := &Blockchain{tip, db}

  return bc
}

func (bc *Blockchain) MineBlock(transactions []*Transaction) {
  var lastHash []byte

  bc.db.View(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte(blocksBucket))
    lastHash = b.Get([]byte("l"))

    return nil
  })

  newBlock := NewBlock(transactions, lastHash)

  bc.db.Update(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte(blocksBucket))
    b.Put(newBlock.Hash, newBlock.Serialize())
    b.Put([]byte("l"), newBlock.Hash)
    bc.tip = newBlock.Hash

    return nil
  })
}

func (bc *Blockchain) CloseDB() {
  bc.db.Close()
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
  bci := &BlockchainIterator{bc.tip, bc.db}

  return bci
}

func (i *BlockchainIterator) Next() *Block {
  var block *Block

  i.db.View(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte(blocksBucket))
    encodedBlock := b.Get(i.currentHash)
    block = DeserializeBlock(encodedBlock)

    return nil
  })

  i.currentHash = block.PrevBlockHash

  return block
}

func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
  var unspentTXs []Transaction
  spentTXOs := make(map[string][]int)
  bci := bc.Iterator()

  for {
    block := bci.Next()

    for _, tx := range block.Transactions {
      txID := hex.EncodeToString(tx.ID)
    Outputs:
      for outIdx, out := range tx.Vout {
        if spentTXOs[txID] != nil {
          for _, spentOut := range spentTXOs[txID] {
            if spentOut == outIdx {
              continue Outputs
            }
          }
        }

        if out.CanBeUnlockedWith(address) {
          unspentTXs = append(unspentTXs, *tx)
        }
      }

      if tx.IsCoinbase() == false {
        for _, in := range tx.Vin {
          inTxID := hex.EncodeToString(in.Txid)
          spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
        }
      }
    }

    if len(block.PrevBlockHash) == 0 {
      break
    }
  }

  return unspentTXs
}

func (bc *Blockchain) FindUTXO(address string) []TXOutput {
  var UTXOs []TXOutput
  unspentTransactions := bc.FindUnspentTransactions(address)

  for _, unspentTX := range unspentTransactions {
    for _, out := range unspentTX.Vout {
      if out.CanBeUnlockedWith(address) {
        UTXOs = append(UTXOs, out)
      }
    }
  }

  return UTXOs
}

func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
  unspentOutputs := make(map[string][]int)
  unspentTXs := bc.FindUnspentTransactions(address)
  accumulated := 0

Work:
  for _, tx := range unspentTXs {
    txID := hex.EncodeToString(tx.ID)

    for outIdx, out := range tx.Vout {
      if out.CanBeUnlockedWith(address) && accumulated < amount {
        accumulated += out.Value
        unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

        if accumulated >= amount {
          break Work
        }
      }
    }
  }

  return accumulated, unspentOutputs
}
