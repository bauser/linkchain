package main

import (
  "flag"
  "fmt"
  "github.com/bauser/linkchain/core"
  "log"
  "os"
  "strconv"
)

type CLI struct{}

func (cli *CLI) Run() {
  cli.validateArgs()

  createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
  printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
  getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
  sendCmd := flag.NewFlagSet("send", flag.ExitOnError)

  createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
  getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance of")
  sendFrom := sendCmd.String("from", "", "Sender address")
  sendTo := sendCmd.String("to", "", "Receiver address")
  sendAmount := sendCmd.String("amount", "", "Amount to send")

  switch os.Args[1] {
  case "createblockchain":
    err := createBlockchainCmd.Parse(os.Args[2:])
    if err != nil {
      log.Panic(err)
    }
  case "printchain":
    err := printChainCmd.Parse(os.Args[2:])
    if err != nil {
      log.Panic(err)
    }
  case "getbalance":
    err := getBalanceCmd.Parse(os.Args[2:])
    if err != nil {
      log.Panic(err)
    }
  case "send":
    err := sendCmd.Parse(os.Args[2:])
    if err != nil {
      log.Panic(err)
    }
  default:
    cli.printUsage()
    os.Exit(1)
  }

  if createBlockchainCmd.Parsed() {
    if *createBlockchainAddress == "" {
      createBlockchainCmd.Usage()
      os.Exit(1)
    } else {
      cli.createBlockchain(*createBlockchainAddress)
    }
  }

  if printChainCmd.Parsed() {
    cli.printChain()
  }

  if getBalanceCmd.Parsed() {
    if *getBalanceAddress == "" {
      getBalanceCmd.Usage()
      os.Exit(1)
    } else {
      cli.getBalance(*getBalanceAddress)
    }
  }

  if sendCmd.Parsed() {
    amount, err := strconv.Atoi(*sendAmount)
    if err != nil {
      log.Panic(err)
    }
    if *sendFrom == "" || *sendTo == "" || amount <= 0 {
      sendCmd.Usage()
      os.Exit(1)
    } else {
      cli.send(*sendFrom, *sendTo, amount)
    }
  }
}

func (cli *CLI) createBlockchain(address string) {
  bc := core.NewBlockchain(address)
  bc.CloseDB()

  fmt.Println("Done!")
}

func (cli *CLI) printChain() {
  bc := core.NewBlockchain("")
  defer bc.CloseDB()

  bci := bc.Iterator()

  for {
    block := bci.Next()

    fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
    fmt.Printf("Hash: %x\n", block.Hash)
    pow := core.NewProofOfWork(block)
    fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
    fmt.Println()

    if len(block.PrevBlockHash) == 0 {
      break
    }
  }
}

func (cli *CLI) getBalance(address string) {
  bc := core.NewBlockchain(address)
  defer bc.CloseDB()

  balance := 0
  UTXOs := bc.FindUTXO(address)

  for _, out := range UTXOs {
    balance += out.Value
  }

  fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) send(from, to string, amount int) {
  bc := core.NewBlockchain(from)
  defer bc.CloseDB()

  tx := core.NewUTXOTransaction(from, to, amount, bc)
  bc.MineBlock([]*core.Transaction{tx})
  fmt.Println("Success!")
}

func (cli *CLI) validateArgs() {

}

func (cli *CLI) printUsage() {

}
