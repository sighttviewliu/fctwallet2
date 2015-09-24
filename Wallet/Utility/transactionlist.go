// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package Utility


import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/factoid/block"
	fct "github.com/FactomProject/factoid"
	"github.com/FactomProject/factom"
	"github.com/FactomProject/FactomCode/common"
)

/************************************************
 * Transaction listing code
 ***********************************************/

// Older blocks smaller indexes.  All the Factoid Directory blocks
var DirectoryBlocks  = make([]*common.DirectoryBlock,0,100)
var FactoidBlocks    = make([]block.IFBlock,0,100)
var DBHead    []byte = common.ZERO_HASH
var DBHeadStr string = ""
var DBHeadLast []byte = common.ZERO_HASH	
	
// Refresh the Directory Block Head.  If it has changed, return true.
// Otherwise return false.
func getDBHead() bool {
	db, err := factom.GetDBlockHead()
	
	if err != nil {
		panic(err.Error())
	}
	
	if db.KeyMR != DBHeadStr {
		DBHeadStr = db.KeyMR
		DBHead,err = hex.DecodeString(db.KeyMR)
		if err != nil {
			panic(err.Error())
		}
		
		return true
	}
	return false
}

func getAll() error {
	dbs := make([] *common.DirectoryBlock,0,100)
	next := DBHeadStr
	
	for {
		blk,err := factom.GetRaw(next)
		if err != nil {
			panic(err.Error())
		}
		db := new(common.DirectoryBlock)
		err = db.UnmarshalBinary(blk)
		if err != nil {
			panic(err.Error())
		}
		dbs = append(dbs,db)
		if bytes.Equal(db.Header.PrevKeyMR.Bytes(),DBHeadLast) {
			break
		}
		next = hex.EncodeToString(db.Header.PrevKeyMR.Bytes())
	}
	
	DBHeadLast = DBHead
		
	for i:= len(dbs)-1;i>=0; i-- {
		DirectoryBlocks = append(DirectoryBlocks,dbs[i])
		fb := new(block.FBlock)
		var fcnt int
		for _,dbe := range dbs[i].DBEntries {
			if bytes.Equal(dbe.ChainID.Bytes(),common.FACTOID_CHAINID) {
				fcnt++
				hashstr := hex.EncodeToString(dbe.KeyMR.Bytes())
				fdata,err := factom.GetRaw(hashstr)
				if err != nil {
					panic(err.Error())
				}
				err = fb.UnmarshalBinary(fdata)
				if err != nil {
					panic(err.Error())
				}
				FactoidBlocks = append(FactoidBlocks,fb)
				break
			}
		}
		if fb == nil {
			panic("Missing Factoid Block from a directory block")
		}
		if fcnt > 1 {
			panic("More than one Factom Block found in a directory block.")
		}
		if err := ProcessFB(fb); err != nil {
			return err
		}
	}
	return nil
}

func refresh() error {

	if getDBHead() {
		if err := getAll(); err != nil {
			return err
		}
	}
	return nil
}

func filtertransaction(trans fct.ITransaction, addresses [][]byte) bool {
	if addresses == nil || len(addresses)==0 {
		return true
	}
	for _,adr := range addresses {
		for _,in := range trans.GetInputs() {
			if bytes.Equal(adr,in.GetAddress().Bytes()) {
				return true
			}
		}
		for _,out := range trans.GetOutputs() {
			if bytes.Equal(adr,out.GetAddress().Bytes()) {
				return true
			}
		}
		for _,ec := range trans.GetECOutputs() {
			if bytes.Equal(adr,ec.GetAddress().Bytes()) {
				return true
			}
		}
	}
	return false
}

func DumpTransactions(addresses [][]byte) ([]byte, error) {
	var ret bytes.Buffer
	if err := refresh(); err != nil {
		return nil, err
	}
	transcnt := 1
	for i,fb := range FactoidBlocks {
		var out bytes.Buffer
		wrt := false
		if len(fb.GetTransactions()) > 1 {
			
			out.WriteString(fmt.Sprintf("Transactions at block height %d\n",i))
			for j, t := range fb.GetTransactions() {
				var _ = j
				if j != 0 {
					if filtertransaction(t,addresses) {
						out.WriteString(fmt.Sprintf("Transaction %d\n",transcnt))
						out.WriteString(fmt.Sprintf("%s\n",t.String()))
						wrt = true
					}
					transcnt++
				}
			}
		}
		if wrt {
			ret.WriteString(out.String())
		}
	}
	return ret.Bytes(), nil
}

// At some point we will need to be smarter... Process Blocks and transactions here!
func ProcessFB(fb block.IFBlock) error {
	return nil
}