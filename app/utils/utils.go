package utils

import (
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

func PopulateWallet(wallet *gateway.Wallet, org string) error {
	log.Println("============ Populating wallet ============")
	orgPath := fmt.Sprintf("%s.example.com", org)
	usrPath := fmt.Sprintf("User1@%s.example.com", org)
	orgMSP := strings.ToUpper(org[:1]) + org[1:]

	credPath := filepath.Join(
		"..",
		"infrastructure",
		"organizations",
		"peerOrganizations",
		orgPath,
		"users",
		usrPath,
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

	identity := gateway.NewX509Identity(fmt.Sprintf("%sMSP", orgMSP), string(cert), string(key))

	return wallet.Put("usr1", identity)
}

func CreateWallet(userId, userOrg string) (*gateway.Wallet, error) {
	walletPath := fmt.Sprintf("wallet/%s", userOrg)
	wallet, err := gateway.NewFileSystemWallet(walletPath)
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
		return nil, err
	}

	if !wallet.Exists(userId) {
		err = PopulateWallet(wallet, userOrg)
		if err != nil {
			log.Fatalf("Failed to populate wallet contents: %v", err)
			return nil, err
		}
	}

	return wallet, nil
}

func ConnectToGateway(wallet *gateway.Wallet, org string) (*gateway.Gateway, error) {
	orgPath := fmt.Sprintf("%s.example.com", org)
	connection := fmt.Sprintf("connection-%s.json", org)
	ccpPath := filepath.Join(
		"..",
		"infrastructure",
		"organizations",
		"peerOrganizations",
		orgPath,
		connection,
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "usr1"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
		return nil, err
	}

	return gw, nil
}
