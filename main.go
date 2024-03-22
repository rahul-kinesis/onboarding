package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {

	// welcome
	fmt.Println("Welcome to Kinesis network onboarding")

	// Load env variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Read variables from .env file
	rpcURL := os.Getenv("RPC_URL")
	privateKeyHex := os.Getenv("PRIVATE_KEY")
	gasLimitStr := os.Getenv("GAS_LIMIT")
	onboardingABI := os.Getenv("ONBOARDING_ABI")

	// Convert gasLimitStr to uint64
	gasLimit, err := strconv.ParseUint(gasLimitStr, 10, 64)
	if err != nil {
		log.Fatalf("Failed to parse gas limit: %v", err)
	}

	// Connect to the Ethereum node
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	// Load the ABI (Application Binary Interface) of the smart contract
	contractAbi, err := abi.JSON(strings.NewReader(onboardingABI))
	if err != nil {
		log.Fatalf("Failed to parse ABI: %v", err)
	}

	// Load the private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	// Deploy the contract
	nonce, err := client.PendingNonceAt(context.Background(), crypto.PubkeyToAddress(privateKey.PublicKey))
	if err != nil {
		log.Fatalf("Failed to retrieve nonce: %v", err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("Failed to retrieve gas price: %v", err)
	}

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = gasLimit   // in units
	auth.GasPrice = gasPrice   // in wei

	address, _, err := deployContract(client, auth, contractAbi)
	if err != nil {
		log.Fatalf("Failed to deploy contract: %v", err)
	}

	fmt.Printf("Contract deployed at address: %s\n", address.Hex())

	// Increase gas price by 10% for safer writes
	newGasPrice := new(big.Int).Mul(gasPrice, big.NewInt(110))
	newGasPrice = newGasPrice.Div(newGasPrice, big.NewInt(100))
	auth.GasPrice = newGasPrice

	// Interact with the contract
	instance, err := NewMain(address, client)

	// Write to the contract
	tx, err := instance.SetMessage(auth, "New vow")
	if err != nil {
		log.Fatalf("Failed to write to contract: %v", err)
	}
	fmt.Printf("Transaction sent: %s\n", tx.Hash().Hex())

	ReadFromTransaction(client, tx.Hash())

}

// Function to deploy the contract
func deployContract(client *ethclient.Client, auth *bind.TransactOpts, contractAbi abi.ABI) (common.Address, *types.Transaction, error) {
	address, tx, _, err := bind.DeployContract(auth, contractAbi, []byte{}, client)
	if err != nil {
		return common.Address{}, nil, fmt.Errorf("Failed to deploy contract: %v", err)
	}
	return address, tx, nil
}

func ReadFromTransaction(client *ethclient.Client, txHash common.Hash) {
	fmt.Println("waiting...")
	time.Sleep(30 * time.Second)

	// Get transaction receipt
	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		log.Fatalf("Failed to get transaction receipt: %v", err)
	}

	// Parse all logs
	for _, log := range receipt.Logs {
		// Attempt to interpret log data as a string
		str := string(log.Data)
		fmt.Println("Potential string written in transaction:", str)
	}

}
