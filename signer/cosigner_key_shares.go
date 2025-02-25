package signer

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"os"

	cometjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/privval"
	ecies "github.com/ecies/go/v2"
	tsed25519 "gitlab.com/unit410/threshold-ed25519/pkg"
	"golang.org/x/sync/errgroup"
)

// CreateCosignerEd25519ShardsFromFile creates CosignerEd25519Key objects from a priv_validator_key.json file
func CreateCosignerEd25519ShardsFromFile(priv string, threshold, shards uint8) ([]CosignerEd25519Key, error) {
	pv, err := ReadPrivValidatorFile(priv)
	if err != nil {
		return nil, err
	}
	return CreateCosignerEd25519Shards(pv, threshold, shards), nil
}

// CreateCosignerEd25519Shards creates CosignerEd25519Key objects from a privval.FilePVKey
func CreateCosignerEd25519Shards(pv privval.FilePVKey, threshold, shards uint8) []CosignerEd25519Key {
	privShards := tsed25519.DealShares(tsed25519.ExpandSecret(pv.PrivKey.Bytes()[:32]), threshold, shards)
	out := make([]CosignerEd25519Key, shards)
	for i, shard := range privShards {
		out[i] = CosignerEd25519Key{
			PubKey:       pv.PubKey,
			PrivateShard: shard,
			ID:           i + 1,
		}
	}
	return out
}

// CreateCosignerRSAShards generate  CosignerRSAKey objects.
func CreateCosignerRSAShards(shards int) ([]CosignerRSAKey, error) {
	rsaKeys, pubKeys, err := makeRSAKeys(shards)
	if err != nil {
		return nil, err
	}
	out := make([]CosignerRSAKey, shards)
	for i, key := range rsaKeys {
		out[i] = CosignerRSAKey{
			ID:      i + 1,
			RSAKey:  *key,
			RSAPubs: pubKeys,
		}
	}
	return out, nil
}

// ReadPrivValidatorFile reads in a privval.FilePVKey from a given file.
func ReadPrivValidatorFile(priv string) (out privval.FilePVKey, err error) {
	var bz []byte
	if bz, err = os.ReadFile(priv); err != nil {
		return
	}
	if err = cometjson.Unmarshal(bz, &out); err != nil {
		return
	}
	return
}

// WriteCosignerEd25519ShardFile writes a cosigner Ed25519 key to a given file name.
func WriteCosignerEd25519ShardFile(cosigner CosignerEd25519Key, file string) error {
	jsonBytes, err := json.Marshal(&cosigner)
	if err != nil {
		return err
	}
	return os.WriteFile(file, jsonBytes, 0600)
}

// WriteCosignerRSAShardFile writes a cosigner RSA key to a given file name.
func WriteCosignerRSAShardFile(cosigner CosignerRSAKey, file string) error {
	jsonBytes, err := json.Marshal(&cosigner)
	if err != nil {
		return err
	}
	return os.WriteFile(file, jsonBytes, 0600)
}

// CreateCosignerECIESShards generates CosignerECIESKey objects.
func CreateCosignerECIESShards(shards int) ([]CosignerECIESKey, error) {
	eciesKeys, pubKeys, err := makeECIESKeys(shards)
	if err != nil {
		return nil, err
	}
	out := make([]CosignerECIESKey, shards)
	for i, key := range eciesKeys {
		out[i] = CosignerECIESKey{
			ID:        i + 1,
			ECIESKey:  key,
			ECIESPubs: pubKeys,
		}
	}
	return out, nil
}

// WriteCosignerECIESShardFile writes a cosigner ECIES key to a given file name.
func WriteCosignerECIESShardFile(cosigner CosignerECIESKey, file string) error {
	jsonBytes, err := json.Marshal(&cosigner)
	if err != nil {
		return err
	}
	return os.WriteFile(file, jsonBytes, 0600)
}

func makeRSAKeys(num int) (rsaKeys []*rsa.PrivateKey, pubKeys []*rsa.PublicKey, err error) {
	rsaKeys = make([]*rsa.PrivateKey, num)
	pubKeys = make([]*rsa.PublicKey, num)
	var eg errgroup.Group
	bitSize := 4096
	for i := 0; i < num; i++ {
		i := i
		eg.Go(func() error {
			rsaKey, err := rsa.GenerateKey(rand.Reader, bitSize)
			if err != nil {
				return err
			}
			rsaKeys[i] = rsaKey
			pubKeys[i] = &rsaKey.PublicKey

			return nil
		})
	}
	return rsaKeys, pubKeys, eg.Wait()
}

func makeECIESKeys(num int) ([]*ecies.PrivateKey, []*ecies.PublicKey, error) {
	eciesKeys := make([]*ecies.PrivateKey, num)
	pubKeys := make([]*ecies.PublicKey, num)
	var eg errgroup.Group
	for i := 0; i < num; i++ {
		i := i
		eg.Go(func() error {
			eciesKey, err := ecies.GenerateKey()
			if err != nil {
				return err
			}
			eciesKeys[i] = eciesKey
			pubKeys[i] = eciesKey.PublicKey
			return nil
		})
	}
	return eciesKeys, pubKeys, eg.Wait()
}
