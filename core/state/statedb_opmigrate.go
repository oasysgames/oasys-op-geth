package state

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
)

// StorageTrie returns the storage trie of an account. The return value is a copy
// and is nil for non-existent accounts. An error will be returned if storage trie
// is existent but can't be loaded correctly.
func (s *StateDB) StorageTrie(addr common.Address) (Trie, error) {
	stateObject := s.getStateObject(addr)
	if stateObject == nil {
		return nil, nil
	}
	cpy := stateObject.deepCopy(s)
	if _, err := cpy.updateTrie(); err != nil {
		return nil, err
	}
	return cpy.getTrie()
}

func (db *StateDB) ForEachStorage(addr common.Address, cb func(key, value common.Hash) bool) error {
	so := db.getStateObject(addr)
	if so == nil {
		return nil
	}
	tr, err := so.getTrie()
	if err != nil {
		return err
	}
	nit, err := tr.NodeIterator(nil)
	if err != nil {
		return err
	}
	it := trie.NewIterator(nit)

	for it.Next() {
		key := common.BytesToHash(db.trie.GetKey(it.Key))
		if value, dirty := so.dirtyStorage[key]; dirty {
			if !cb(key, value) {
				return nil
			}
			continue
		}

		if len(it.Value) > 0 {
			_, content, _, err := rlp.Split(it.Value)
			if err != nil {
				return err
			}
			if !cb(key, common.BytesToHash(content)) {
				return nil
			}
		}
	}
	return nil
}
