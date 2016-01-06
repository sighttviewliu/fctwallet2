// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package Wallet

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	fct "github.com/FactomProject/factoid"
	"github.com/FactomProject/factoid/wallet"
	"github.com/FactomProject/fctwallet2/Wallet/Utility"
)

func CommitChain(name string, data []byte) error {
	type walletcommit struct {
		Message string
	}

	type commit struct {
		CommitChainMsg string
	}

	in := new(walletcommit)
	json.Unmarshal(data, in)
	msg, err := hex.DecodeString(in.Message)
	if err != nil {
		return fmt.Errorf("Could not decode message:", err)
	}

	var we fct.IBlock
	
	if Utility.IsValidAddress(name) && strings.HasPrefix(name,"EC") {
		addr := fct.ConvertUserStrToAddress(name)
		we = factoidState.GetDB().GetRaw([]byte(fct.W_ADDRESS_PUB_KEY),addr)
	}else if Utility.IsValidHexAddress(name) {
		addr, err := hex.DecodeString(name)
		if err == nil {
			we = factoidState.GetDB().GetRaw([]byte(fct.W_ADDRESS_PUB_KEY),addr)
		}
	}else{
		we = factoidState.GetDB().GetRaw([]byte(fct.W_NAME), []byte(name))
	}
	
	if we == nil {
		return fmt.Errorf("Unknown address")
	}
	
	signed := factoidState.GetWallet().SignCommit(we.(wallet.IWalletEntry), msg)

	com := new(commit)
	com.CommitChainMsg = hex.EncodeToString(signed)
	j, err := json.Marshal(com)
	if err != nil {
		return fmt.Errorf("Could not create json post:", err)
	}

	resp, err := http.Post(
		fmt.Sprintf("http://%s/v1/commit-chain", ipaddressFD+portNumberFD),
		"application/json",
		bytes.NewBuffer(j))
	if err != nil {
		return fmt.Errorf("Could not post to server:", err)
	}
	resp.Body.Close()

	return nil
}

func CommitEntry(name string, data []byte) error {
	type walletcommit struct {
		Message string
	}

	type commit struct {
		CommitEntryMsg string
	}

	in := new(walletcommit)
	json.Unmarshal(data, in)
	msg, err := hex.DecodeString(in.Message)
	if err != nil {
		return fmt.Errorf("Could not decode message:", err)
	}

	we := factoidState.GetDB().GetRaw([]byte(fct.W_NAME), []byte(name))
	signed := factoidState.GetWallet().SignCommit(we.(wallet.IWalletEntry), msg)

	com := new(commit)
	com.CommitEntryMsg = hex.EncodeToString(signed)
	j, err := json.Marshal(com)
	if err != nil {
		return fmt.Errorf("Could not create json post:", err)
	}

	resp, err := http.Post(
		fmt.Sprintf("http://%s/v1/commit-entry/", ipaddressFD+portNumberFD),
		"application/json",
		bytes.NewBuffer(j))
	if err != nil {
		return fmt.Errorf("Could not post to server:", err)
	}
	resp.Body.Close()
	return nil
}
