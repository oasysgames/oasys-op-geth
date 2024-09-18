package core

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/log"
)

var (
	contractUpdateConfig = make(map[uint64]GenesisAlloc)
)

/*
Load the contract update configuration file(json format).

Example:

	{
	  "100": {
	    "0xfC76559Ffd6EF3b79C2A8Ab1A8179134d1e88953": {
		  "balance": "0",
		  "code": "0x73000000000000000000000000000000000000000030146080604052600080fdfea2646970667358221220184afaf8d58620cf6631ca9af99040ca3677ef4e6257938d587162f4e5edb88664736f6c63430008090033",
		  "storage": {
	        "0x0000000000000000000000000000000000000000000000000000000000000001": "0x000000000000000000000000464110713EAF4E7834D93E68a23B2aD8cCd7b28B",
	        "0x0000000000000000000000000000000000000000000000000000000000000002": "0x00000000000000000000000059BCA8bFfB73900261012c7c72515ecD2e529c97"
		  }
	    }
	  }
	}
*/
func LoadContractUpdateConfig(filepath string) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		panic(fmt.Sprintf("Failed to read the contract update configuration file: %s", err.Error()))
	}

	parsed := make(map[string]GenesisAlloc)
	if err = json.Unmarshal(data, &parsed); err != nil {
		panic(fmt.Sprintf("Failed to unmarshal the contract update configuration file: %s", err.Error()))
	}

	for block, alloc := range parsed {
		if u64block, err := strconv.ParseUint(block, 10, 64); err == nil {
			contractUpdateConfig[u64block] = alloc
		} else {
			panic(fmt.Sprintf("Failed to parse block number: %s, err: %s", block, err.Error()))
		}
	}

	log.Info("Loaded contract update config", "file", filepath)
}

// Update the contract directly to State. Affects the state root of the block.
func UpdateContract(state *state.StateDB, block uint64, on string) {
	allocs, ok := contractUpdateConfig[block]
	if !ok {
		return
	}

	for address, alloc := range allocs {
		if alloc.Code != nil {
			oldHash := state.GetCodeHash(address)
			state.SetCode(address, alloc.Code)
			newHash := state.GetCodeHash(address)

			log.Info("Updated contract code", "on", on,
				"block", block, "address", address,
				"old-hash", oldHash.Hex(), "new-hash", newHash.Hex())
		}

		for slot, value := range alloc.Storage {
			oldValue := state.GetState(address, slot)
			state.SetState(address, slot, value)

			log.Info("Updated contract storage", "on", on,
				"block", block, "address", address,
				"slot", slot, "old-value", oldValue.Hex(), "new-value", value.Hex())
		}
	}
}
