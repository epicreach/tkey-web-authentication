package tkey

import (
	"crypto"
	"crypto/ed25519"
	"fmt"
	"log"
	"os"

	"github.com/tillitis/tkeyclient"
)

const progname = "tkey-device-signer"

var le = log.New(os.Stderr, "Error: ", 0)
var existingSigner *Signer

func GetTkeyPubKey() (ed25519.PublicKey, error) {
	signer, err := getSigner()

	if err != nil {
		return nil, err
	}

	if !signer.connect() {
		le.Printf("Connect failed")
		return nil, fmt.Errorf("connect failed")
	}

	defer signer.disconnect()

	pub, err := signer.tkSigner.GetPubkey()

	if err != nil {
		return nil, err
	}

	pubkey := ed25519.PublicKey(pub)

	signer.printAuthorizedKey()

	return pubkey, nil
}

func Sign(msg []byte) ([]byte, error) {

	signer, err := getSigner()

	if err != nil {
		return nil, err
	}

	if !signer.connect() {
		le.Printf("Connect failed")
		return nil, fmt.Errorf("connect failed")
	}

	defer signer.disconnect()

	sig, err := signer.Sign(nil, msg, crypto.Hash(0))
	if err != nil {
		le.Printf("Sign failed: %s\n", err)
		return nil, err
	}

	return sig, nil
}

func getSigner() (*Signer, error) {
	if existingSigner != nil && existingSigner.connect() && existingSigner.isWantedApp() {
		// The signer app is already loaded, return the existing signer
		return existingSigner, nil
	}

	devPath, err := tkeyclient.DetectSerialPort(false)
	if err != nil {
		return nil, err
	}

	serialSpeed := tkeyclient.SerialSpeed

	exit := func(code int) {
		os.Exit(0)
	}

	enterUSS := true
	fileUSS := ""

	signer := NewSigner(devPath, serialSpeed, enterUSS, fileUSS, "", exit)
	existingSigner = signer

	return signer, nil
}
