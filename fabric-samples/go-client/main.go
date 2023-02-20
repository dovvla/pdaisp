/*
Copyright 2021 IBM All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

func formatJSON(data []byte) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", ""); err != nil {
		panic(fmt.Errorf("failed to parse JSON: %w", err))
	}
	return prettyJSON.String()

}

var orgNum string

func populateWallet(wallet *gateway.Wallet) error {

	credPath := filepath.Join(
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		orgNum+".example.com",
		"users",
		"User1@"+orgNum+".example.com",
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return errors.New("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity(strings.Title(orgNum)+"MSP", string(cert), string(key))

	err = wallet.Put("User1@"+orgNum+".example.com", identity)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	orgNum = "org1"
	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		fmt.Printf("Failed to create wallet: %s\n", err)
		os.Exit(1)
	}

	if !wallet.Exists("User1@" + orgNum + ".example.com") {
		err = populateWallet(wallet)
		if err != nil {
			fmt.Printf("Failed to populate wallet contents: %s\n", err)
			os.Exit(1)
		}
	}

	channelName := "mychannel"
	chaincodeName := "basic"

	ccpPath := filepath.Join(
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		orgNum+".example.com",
		"connection-"+orgNum+".json",
	)
	gateway, err := gateway.Connect(
		gateway.WithIdentity(wallet, "User1@"+orgNum+".example.com"),

		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
	)
	if err != nil {
		panic(err)
	}
	defer gateway.Close()

	network, err := gateway.GetNetwork(channelName)

	if err != nil {
		panic(err)
	}

	contract := network.GetContract(chaincodeName)

	var option int

loop:
	for {
		fmt.Println("Menu, choose option: ")
		fmt.Println("0: Initialize ledger")
		fmt.Println("1: Read person asset")
		fmt.Println("2: Read car asset")
		fmt.Println("3: Get cars by color")
		fmt.Println("4: Get cars by color and owner")
		fmt.Println("5: Transfer car to another owner")
		fmt.Println("6: Add car fault")
		fmt.Println("7: Change car color")
		fmt.Println("8: Repair car")
		fmt.Println("9: Exit")

		fmt.Scanf("%d", &option)

		switch option {
		case 0:
			fmt.Println("Initializing ledger...")
			initLedger(contract)

		case 1:
			fmt.Printf("Person ID: ")
			var personID string
			fmt.Scanf("%s", &personID)
			readPersonAsset(contract, personID)

		case 2:
			fmt.Printf("Car ID: ")
			var carID string
			fmt.Scanf("%s", &carID)
			readCarAsset(contract, carID)

		case 3:
			fmt.Printf("Car color: ")
			var color string
			fmt.Scanf("%s", &color)
			getCarsByColor(contract, color)

		case 4:
			fmt.Printf("Car color: ")
			var color string
			fmt.Scanf("%s", &color)

			fmt.Printf("Car owner: ")
			var ownerID string
			fmt.Scanf("%s", &ownerID)
			getCarsByColorAndOwner(contract, color, ownerID)

		case 5:
			fmt.Printf("Car ID: ")
			var carID string
			fmt.Scanf("%s", &carID)

			fmt.Printf("New owner ID: ")
			var newOwnerID string
			fmt.Scanf("%s", &newOwnerID)

			fmt.Printf("Accept faulted car? (Y/n): ")
			var accept string
			fmt.Scanf("%s", &accept)
			var acceptfaultedBool bool
			if accept == "n" {
				acceptfaultedBool = false
			} else {
				acceptfaultedBool = true
			}

			transferCarAsset(contract, carID, newOwnerID, acceptfaultedBool)

		case 6:
			fmt.Printf("Car ID: ")
			var carID string
			fmt.Scanf("%s", &carID)

			fmt.Println("Fault description:")
			var description string
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				description = scanner.Text()
			}

			fmt.Printf("Fault repair price: ")
			var repairPrice float32
			fmt.Scanf("%f", &repairPrice)
			addCarfault(contract, carID, description, repairPrice)

		case 7:
			fmt.Printf("Car ID: ")
			var carID string
			fmt.Scanf("%s", &carID)

			fmt.Printf("New car color: ")
			var newColor string
			fmt.Scanf("%s", &newColor)
			changeCarColor(contract, carID, newColor)

		case 8:
			fmt.Printf("Car ID: ")
			var carID string
			fmt.Scanf("%s", &carID)
			repairCar(contract, carID)

		case 9:
			break loop

		default:
			fmt.Printf("No such option")
		}

		fmt.Printf("\n")
	}

}

func initLedger(contract *gateway.Contract) {
	_, err := contract.SubmitTransaction("InitLedger")
	if err != nil {
		fmt.Println(fmt.Errorf("failed to submit transaction: %w", err))
		return
	}

	fmt.Printf("######## Success ########\n")
}

func readPersonAsset(contract *gateway.Contract, id string) {
	evaluateResult, err := contract.EvaluateTransaction("ReadPersonAsset", id)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to evaluate transaction: %w", err))
		return
	}
	result := formatJSON(evaluateResult)

	fmt.Printf("######## Result:%s\n", result)
}

func readCarAsset(contract *gateway.Contract, id string) {
	evaluateResult, err := contract.EvaluateTransaction("ReadCarAsset", id)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to evaluate transaction: %w", err))
		return
	}
	result := formatJSON(evaluateResult)

	fmt.Printf("######## Result:%s\n", result)
}

func getCarsByColor(contract *gateway.Contract, color string) {
	evaluateResult, err := contract.EvaluateTransaction("GetCarsByColor", color)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to evaluate transaction: %w", err))
		return
	}
	result := formatJSON(evaluateResult)

	fmt.Printf("######## Result:%s\n", result)
}

func getCarsByColorAndOwner(contract *gateway.Contract, color string, ownerID string) {

	evaluateResult, err := contract.EvaluateTransaction("GetCarsByColorAndOwner", color, ownerID)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to evaluate transaction: %w", err))
		return
	}
	result := formatJSON(evaluateResult)

	fmt.Printf("######## Result:%s\n", result)
}

func transferCarAsset(contract *gateway.Contract, id string, newOwner string, acceptfault bool) {

	_, err := contract.SubmitTransaction("TransferCarAsset", id, newOwner, strconv.FormatBool(acceptfault))
	if err != nil {
		fmt.Println(fmt.Errorf("failed to submit transaction: %w", err))
		return
	}
	fmt.Printf("######## Success ########\n")

}

func addCarfault(contract *gateway.Contract, id string, description string, repairPrice float32) {

	_, err := contract.SubmitTransaction("AddCarfault", id, description, fmt.Sprintf("%f", repairPrice))
	if err != nil {
		fmt.Println(fmt.Errorf("failed to submit transaction: %w", err))
		return
	}
	fmt.Printf("######## Success ########\n")

}

func changeCarColor(contract *gateway.Contract, id string, newColor string) {

	_, err := contract.SubmitTransaction("ChangeCarColor", id, newColor)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to submit transaction: %w", err))
		return
	}

	fmt.Printf("######## Success ########\n")
}

func repairCar(contract *gateway.Contract, id string) {

	_, err := contract.SubmitTransaction("RepairCar", id)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to submit transaction: %w", err))
		return
	}

	fmt.Printf("######## Success ########\n")
}
