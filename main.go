package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"
)

var rpcURL string
var privateKeyHex string
var gasLimitStr string
var onboardingABI string
var mongoURL string

func main() {

	// welcome
	fmt.Println("Welcome to Kinesis network onboarding")

	// Load env variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Read variables from .env file
	initialize()

	// Convert gasLimitStr to uint64
	_, err = strconv.ParseUint(gasLimitStr, 10, 64)
	if err != nil {
		log.Fatalf("Failed to parse gas limit: %v", err)
	}

	// Connect to the Ethereum node
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	//
	//// Load the ABI (Application Binary Interface) of the smart contract
	//contractAbi, err := abi.JSON(strings.NewReader(onboardingABI))
	//if err != nil {
	//	log.Fatalf("Failed to parse ABI: %v", err)
	//}
	//
	//// Load the private key
	//privateKey, err := crypto.HexToECDSA(privateKeyHex)
	//if err != nil {
	//	log.Fatalf("Failed to load private key: %v", err)
	//}
	//
	//// Deploy the contract
	//nonce, err := client.PendingNonceAt(context.Background(), crypto.PubkeyToAddress(privateKey.PublicKey))
	//if err != nil {
	//	log.Fatalf("Failed to retrieve nonce: %v", err)
	//}
	//
	//gasPrice, err := client.SuggestGasPrice(context.Background())
	//if err != nil {
	//	log.Fatalf("Failed to retrieve gas price: %v", err)
	//}
	//
	//auth := bind.NewKeyedTransactor(privateKey)
	//auth.Nonce = big.NewInt(int64(nonce))
	//auth.Value = big.NewInt(0) // in wei
	//auth.GasLimit = gasLimit   // in units
	//auth.GasPrice = gasPrice   // in wei
	//
	//address, _, err := deployContract(client, auth, contractAbi)
	//if err != nil {
	//	log.Fatalf("Failed to deploy contract: %v", err)
	//}
	//
	//fmt.Printf("Contract deployed at address: %s\n", address.Hex())
	//
	//// Increase gas price by 10% for safer writes
	//newGasPrice := new(big.Int).Mul(gasPrice, big.NewInt(110))
	//newGasPrice = newGasPrice.Div(newGasPrice, big.NewInt(100))
	//auth.GasPrice = newGasPrice
	//
	//// Interact with the contract
	//instance, err := NewMain(address, client)
	//
	//// Write to the contract
	//tx, err := instance.SetMessage(auth, "New vow")
	//if err != nil {
	//	log.Fatalf("Failed to write to contract: %v", err)
	//}
	//fmt.Printf("Transaction sent: %s\n", tx.Hash().Hex())
	//
	////err = ReadFromLastThreeDays(client)
	////if err != nil {
	////	fmt.Printf("error fetching last three days logs", err)
	////}

	ReadFromTransaction(client)

}

// initialize variables
func initialize() {
	// Read variables from .env file
	rpcURL = os.Getenv("RPC_URL")
	privateKeyHex = os.Getenv("PRIVATE_KEY")
	gasLimitStr = os.Getenv("GAS_LIMIT")
	onboardingABI = os.Getenv("ONBOARDING_ABI")
	mongoURL = os.Getenv("MONGODB_URI")
}

// Function to deploy the contract
func deployContract(client *ethclient.Client, auth *bind.TransactOpts, contractAbi abi.ABI) (common.Address, *types.Transaction, error) {
	address, tx, _, err := bind.DeployContract(auth, contractAbi, []byte{}, client)
	if err != nil {
		return common.Address{}, nil, fmt.Errorf("Failed to deploy contract: %v", err)
	}
	return address, tx, nil
}

// ReadFromTransaction Function to read from a transaction hash.
// func ReadFromTransaction(client *ethclient.Client, txHash common.Hash, contractAddress common.Address) {
func ReadFromTransaction(client *ethclient.Client) {

	var txHash common.Hash
	var contractAddress common.Address

	txHash = common.HexToHash("0xbb90b230a298ad8e22f4cc58e4c3f65ace32a3ab4e49f7567913c666c62a7e4e")
	contractAddress = common.HexToAddress("0xDA0ED520C966a65567Eeb140fb654971d6ccffcB")

	client, err := ethclient.Dial("https://sepolia.infura.io/v3/8cd437ae269f42ed87fddc11d4fd5a01")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		log.Fatalf("Failed to retrieve transaction receipt: %v", err)
	}

	blockNumber := receipt.BlockNumber

	block, err := client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		log.Fatalf("Failed to retrieve block information: %v", err)
	}

	for _, tx := range block.Transactions() {
		// Check if the transaction is to our contract
		if tx.To() != nil && *tx.To() == contractAddress {
			// Decode transaction data to extract the vow
			vow, err := decodeVowData(tx.Data())
			if err != nil {
				log.Printf("Error decoding vow data: %v", err)
				continue
			}
			fmt.Printf("Vow: %s\n", vow)
			err = saveVowToMongoDB(vow)
			if err != nil {
				log.Fatalf("Failed to save vow to MongoDB: %v", err)
			}
		}
	}
}

// Function to decode transaction data and extract the vow
func decodeVowData(data []byte) (string, error) {
	// Implement decoding logic for your specific smart contract
	// Here, you need to decode the transaction data to extract the vow
	// This may involve parsing the data based on your contract's ABI
	// For simplicity, let's assume the vow is stored as a UTF-8 string

	// Convert bytes to string (assuming vow is stored as string)
	vow := string(data)
	spaceIndex := strings.IndexByte(vow, ' ')
	if spaceIndex == -1 {
		return "", fmt.Errorf("no space character found in vow")
	}

	// Extract the substring starting from the character immediately after the first space
	vow = vow[spaceIndex+1:]

	return vow, nil
	return vow, nil
}

func ReadFromLastThreeDays(client *ethclient.Client) error {
	fmt.Println("Read--")
	// Get the current block number
	threeDaysAgo := time.Now().Add(-72 * time.Hour) // 3 days * 24 hours/day
	blockTime := threeDaysAgo.Unix()

	// Create a filter query to retrieve logs within the last three days
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(blockTime),
		ToBlock:   nil,
		Addresses: []common.Address{}, // Optionally filter by contract addresses
	}

	// Fetch logs matching the filter query
	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Fatalf("Failed to retrieve logs: %v", err)
		return err
	}
	fmt.Printf("logs:%v", logs)
	for _, logg := range logs {
		fmt.Printf("Log Block Number: %d\n", logg.BlockNumber)
		fmt.Printf("Log Address: %s\n", logg.Address.Hex())
		// Add your processing logic here
	}
	return nil
}

func saveVowToMongoDB(vow string) error {
	// Set up MongoDB connection
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().ApplyURI(mongoURL).SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return err
	}
	defer client.Disconnect(context.Background())

	// Access the database and collection
	db := client.Database("kinesis")
	collection := db.Collection("wedding vows")

	// Create a document to insert
	document := bson.M{"vow": vow}

	// Insert document into collection
	_, err = collection.InsertOne(context.Background(), document)
	if err != nil {
		return err
	}

	return nil
}
