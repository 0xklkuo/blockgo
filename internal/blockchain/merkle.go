package blockchain

import "crypto/sha256"

func MerkleRoot(txs []Transaction) Hash {
	if len(txs) == 0 {
		return Hash(sha256.Sum256(nil))
	}

	level := make([]Hash, 0, len(txs))
	for _, tx := range txs {
		level = append(level, tx.ID)
	}

	for len(level) > 1 {
		if len(level)%2 != 0 {
			level = append(level, level[len(level)-1])
		}

		next := make([]Hash, 0, len(level)/2)
		for i := 0; i < len(level); i += 2 {
			combined := append(level[i].Bytes(), level[i+1].Bytes()...)
			sum := sha256.Sum256(combined)
			next = append(next, Hash(sum))
		}

		level = next
	}

	return level[0]
}
