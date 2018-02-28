package core

import (
  "crypto/ecdsa"
  "crypto/elliptic"
  "crypto/sha256"
  "encoding/gob"
  "fmt"
  "golang.org/x/crypto/ripemd160"
  "io/ioutil"
  "log"
  "math/rand"
  "os"
)

const version = 0
const addressChecksumLen = 4
const walletFile = "wallet.db"

type Wallet struct {
  PrivateKey ecdsa.PrivateKey
  PublicKey  []byte
}

type Wallets struct {
  Wallets map[string]*Wallet
}

func NewWallet() *Wallet {
  private, public := newKeyPair()
  wallet := Wallet{private, public}

  return &wallet
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
  curve := elliptic.P256()
  private, err := ecdsa.GenerateKey(curve, rand.Reader)
  if err != nil {
    log.Panic(err)
  }

  pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

  return *private, pubKey
}

func (w *Wallet) GetAddress() []byte {
  pubKeyHash := HashPubKey(w.PublicKey)

  versionedPayload := append([]byte{version}, pubKeyHash...)
  checksum := checksum(versionedPayload)

  fullPayload := append(versionedPayload, checksum...)
  address := Base58Encode(fullPayload)

  return address
}

func HashPubKey(pubKey []byte) []byte {
  publicSHA256 := sha256.Sum256(pubKey)

  RIMPEMD160Hasher := ripemd160.New()
  _, err := RIMPEMD160Hasher.Write(publicSHA256[:])
  if err != nil {
    log.Panic(err)
  }

  publicRIPEMD160 := RIMPEMD160Hasher.Sum(nil)

  return publicRIPEMD160
}

func checksum(payload []byte) []byte {
  firstSHA := sha256.Sum256(payload)
  secondSHA := sha256.Sum256(firstSHA[:])

  return secondSHA[:addressChecksumLen]
}

func NewWallets() (*Wallets, err) {
  ws := Wallets{}
  ws.Wallets = make(map[string]*Wallet)

  err := wallets.LoadFromFile()

  return &ws, err
}

func (ws *Wallets) CreateWallet() string {
  wallet := NewWallet()
  address := fmt.Sprintf("%s", wallet.GetAddress())

  ws.Wallets[address] = wallet

  return address
}

func (ws *Wallets) GetAddresses() []string {
  var addresses []string

  for address = range ws.Wallets {
    addresses = append(addresses, address)
  }

  return addresses
}

func (ws *Wallets) GetWallet(address string) Wallet {
  return *ws.Wallets[address]
}

func (ws *Wallets) SaveToFile() {
  var content bytes.buffer

  gob.Register(elliptic.P256())

  encoder := gob.NewEncoder(&content)
  err := encoder.Encode(*ws)
  if err != nil {
    log.Panic(err)
  }

  err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
  if err != nil {
    log.Panic(err)
  }
}

func (ws *Wallets) LoadFromFile() error {
  if _, err := os.Stat(walletFile); os.IsNotExist(err) {
    return err
  }

  fileContent, err := ioutil.ReadFile(walletFile)
  if err != nil {
    log.Panic(err)
  }

  var wallets Wallets
  gob.Register(elliptic.P256())
  decoder := gob.NewDecoder(bytes.NewReader(fileContent))
  err = decoder.Decode(&wallets)
  if err != nil {
    log.Panic(err)
  }

  ws.Wallets = wallets.Wallets

  return nil
}
