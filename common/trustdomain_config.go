package common

import (
	"crypto/rand"
	"encoding/json"

	"github.com/agl/ed25519"
	"golang.org/x/crypto/nacl/box"
)

// TrustDomainConfig holds the keys for the different talek trust domains.
type TrustDomainConfig struct {
	Name           string
	Address        string
	IsValid        bool
	IsDistributed  bool
	PublicKey      [32]byte // For PIR Encryption
	SignPublicKey  [32]byte // For Signing Interest Vectors
	privateKey     [32]byte
	signPrivateKey [64]byte
}

// PrivateTrustDomainConfig allows export of the trust domain Private Key.
type PrivateTrustDomainConfig struct {
	*TrustDomainConfig
	PrivateKey     [32]byte
	SignPrivateKey [64]byte
}

// NewTrustDomainConfig creates a TrustDomainConfig with a freshly generated keypair.
func NewTrustDomainConfig(name string, address string, isValid bool,
	isDistributed bool) *TrustDomainConfig {
	td := &TrustDomainConfig{}
	td.Name = name
	td.Address = address
	td.IsValid = isValid
	td.IsDistributed = isDistributed
	pubKey, priKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		td.IsValid = false
		return td
	}
	copy(td.PublicKey[:], pubKey[:])
	copy(td.privateKey[:], priKey[:])

	spubKey, spriKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		td.IsValid = false
		return td
	}
	copy(td.SignPublicKey[:], spubKey[:])
	copy(td.signPrivateKey[:], spriKey[:])

	return td
}

// UnmarshalJSON creates a TrustDomainConfig from a serialized form.
func (td *TrustDomainConfig) UnmarshalJSON(marshaled []byte) error {
	if len(marshaled) == 0 {
		return nil
	}
	// The union type between TrustDomainConfig and PrivateTrustDomainConfig.
	type Config struct {
		PublicKey      [32]byte
		PrivateKey     [32]byte
		SignPublicKey  [32]byte
		SignPrivateKey [64]byte
		Name           string
		Address        string
		IsValid        bool
		IsDistributed  bool
	}
	var config Config
	if err := json.Unmarshal(marshaled, &config); err != nil {
		return err
	}

	copy(td.privateKey[:], config.PrivateKey[:])
	copy(td.PublicKey[:], config.PublicKey[:])
	copy(td.signPrivateKey[:], config.SignPrivateKey[:])
	copy(td.SignPublicKey[:], config.SignPublicKey[:])
	td.Name = config.Name
	td.Address = config.Address
	td.IsValid = config.IsValid
	td.IsDistributed = config.IsDistributed

	return nil
}

// Private exposes the Private key of a trust domain config for marshalling.
//   bytes, err := json.Marshal(trustdomainconfig.Private())
func (td *TrustDomainConfig) Private() *PrivateTrustDomainConfig {
	PTDC := new(PrivateTrustDomainConfig)
	PTDC.TrustDomainConfig = td
	copy(PTDC.PrivateKey[:], td.privateKey[:])
	copy(PTDC.SignPrivateKey[:], td.signPrivateKey[:])
	return PTDC
}

// GetName provides the name of the trust domain.
func (td *TrustDomainConfig) GetName() (string, bool) {
	if !td.IsValid {
		return "", false
	}
	return td.Name, td.IsValid
}

// GetAddress returns the remote address of the TrustDomain
func (td *TrustDomainConfig) GetAddress() (string, bool) {
	if !td.IsValid {
		return "", false
	}
	return td.Address, td.IsValid
}
