package utils

import (
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"io/ioutil"
	"log"
	"path/filepath"
)

func PopulateWallet(wallet *gateway.Wallet) error {
	log.Println("============ Populating wallet ============")
	credPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
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
		return fmt.Errorf("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	return wallet.Put("appUser", identity)
}

func CreateWallet(userId, userOrg string) (*gateway.Wallet, error) {
	walletPath := fmt.Sprintf("wallet/%s", userOrg)
	wallet, err := gateway.NewFileSystemWallet(walletPath)
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
		return nil, err
	}

	if !wallet.Exists(userId) {
		err = PopulateWallet(wallet)
		if err != nil {
			log.Fatalf("Failed to populate wallet contents: %v", err)
			return nil, err
		}
	}

	return wallet, nil
}

func ConnectToGateway(wallet *gateway.Wallet) (*gateway.Gateway, error) {
	ccpPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
		return nil, err
	}

	return gw, nil
}
