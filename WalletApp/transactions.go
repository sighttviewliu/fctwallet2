// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	fct "github.com/FactomProject/factoid"
	"github.com/FactomProject/factoid/wallet"
	"net/http"
	"strconv"
)

/************************************************************
 * NewTransaction
 ************************************************************/

type NewTransaction struct {
	ICommand
}

// New Transaction:  key --
// We create a new transaction, and track it with the user supplied key.  The
// user can then use this key to make subsequent calls to add inputs, outputs,
// and to sign. Then they can submit the transaction.
//
// When the transaction is submitted, we clear it from our working memory.
// Multiple transactions can be under construction at one time, but they need
// their own keys. Once a transaction is either submitted or deleted, the key
// can be reused.
func (NewTransaction) Execute(state IState, args []string) error {

	if len(args) != 2 {
		return fmt.Errorf("Invalid Parameters")
	}
	key := args[1]

	// Make sure we don't already have a transaction in process with this key
	t := state.GetFS().GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(key))
	if t != nil {
		return fmt.Errorf("Duplicate key: '%s'", key)
	}
	// Create a transaction
	t = state.GetFS().GetWallet().CreateTransaction(state.GetFS().GetTimeMilli())
	// Save it with the key
	state.GetFS().GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), t)

	fmt.Println("Beginning Transaction ", key)
	return nil
}

func (NewTransaction) Name() string {
	return "NewTransaction"
}

func (NewTransaction) ShortHelp() string {
	return "NewTransaction <key> -- Begins the construction of a transaction.\n" +
		"                        Subsequent modifications must reference the key."
}

func (NewTransaction) LongHelp() string {
	return `
NewTransaction <key>                Begins the construction of a transaction.
                                    The <key> is any token without whitespace up to
                                    32 characters in length that can be used in 
                                    AddInput, AddOutput, AddECOutput, Sign, and
                                    Submit commands to construct and submit 
                                    transactions.
`
}

/************************************************************
 * AddInput
 ************************************************************/

type AddInput struct {
	ICommand
}

// AddInput <key> <name|address> amount
//
//

func (AddInput) Execute(state IState, args []string) error {

	if len(args) != 4 {
		return fmt.Errorf("Invalid Parameters")
	}
	key := args[1]
	adr := args[2]
	amt := args[3]

	ib := state.GetFS().GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(key))
	trans, ok := ib.(fct.ITransaction)
	if ib == nil || !ok {
		return fmt.Errorf("Unknown Transaction")
	}

	var addr fct.IAddress
	if !fct.ValidateFUserStr(adr) {
		if len(adr) != 64 {
			if len(adr) > 32 {
				return fmt.Errorf("Invalid Name.  Name is too long: %v characters", len(adr))
			}

			we := state.GetFS().GetDB().GetRaw([]byte(fct.W_NAME), []byte(adr))

			if we != nil {
				we2 := we.(wallet.IWalletEntry)
				addr, _ = we2.GetAddress()
				adr = hex.EncodeToString(addr.Bytes())
			} else {
				return fmt.Errorf("Name is undefined.")
			}
		} else {
			if badHexChar.FindStringIndex(adr) != nil {
				return fmt.Errorf("Invalid Name.  Name is too long: %v characters", len(adr))
			}
		}
	} else {
		addr = fct.NewAddress(fct.ConvertUserStrToAddress(adr))
	}
	amount, _ := fct.ConvertFixedPoint(amt)
	bamount, _ := strconv.ParseInt(amount, 10, 64)
	err := state.GetFS().GetWallet().AddInput(trans, addr, uint64(bamount))

	if err != nil {
		return err
	}

	fmt.Println("Added Input of ", amt, " to be paid from ", args[2],
		fct.ConvertFctAddressToUserStr(addr))
	return nil
}

func (AddInput) Name() string {
	return "AddInput"
}

func (AddInput) ShortHelp() string {
	return "AddInput <key> <name/address> <amount> -- Adds an input to a transaction.\n" +
		"                              the key should be created by NewTransaction, and\n" +
		"                              and the address and amount should come from your\n" +
		"                              wallet."
}

func (AddInput) LongHelp() string {
	return `
AddInput <key> <name|addr> <amt>    <key>       created by a previous NewTransaction call
                                    <name|addr> A Valid Name in your Factoid Address 
                                                book, or a valid Factoid Address
                                    <amt>       to be sent from the specified address to the
                                                outputs of this transaction.
`
}

/************************************************************
 * AddOutput
 ************************************************************/
type AddOutput struct {
	ICommand
}

// AddOutput <key> <name|address> amount
//
//

func (AddOutput) Execute(state IState, args []string) error {

	if len(args) != 4 {
		return fmt.Errorf("Invalid Parameters")
	}
	key := args[1]
	adr := args[2]
	amt := args[3]

	ib := state.GetFS().GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(key))
	trans, ok := ib.(fct.ITransaction)
	if ib == nil || !ok {
		return fmt.Errorf("Unknown Transaction")
	}

	var addr fct.IAddress
	if !fct.ValidateFUserStr(adr) {
		if len(adr) != 64 {
			if len(adr) > 32 {
				return fmt.Errorf("Invalid Name.  Name is too long: %v characters", len(adr))
			}

			we := state.GetFS().GetDB().GetRaw([]byte(fct.W_NAME), []byte(adr))

			if we != nil {
				we2 := we.(wallet.IWalletEntry)
				addr, _ = we2.GetAddress()
				adr = hex.EncodeToString(addr.Bytes())
			} else {
				return fmt.Errorf("Name is undefined.")
			}
		} else {
			if badHexChar.FindStringIndex(adr) != nil {
				return fmt.Errorf("Invalid Name.  Name is too long: %v characters", len(adr))
			}
		}
	} else {
		addr = fct.NewAddress(fct.ConvertUserStrToAddress(adr))
	}
	amount, _ := fct.ConvertFixedPoint(amt)
	bamount, _ := strconv.ParseInt(amount, 10, 64)
	err := state.GetFS().GetWallet().AddOutput(trans, addr, uint64(bamount))
	if err != nil {
		return err
	}

	fmt.Println("Added Output of ", amt, " to be paid to ", args[2],
		fct.ConvertFctAddressToUserStr(addr))

	return nil
}

func (AddOutput) Name() string {
	return "AddOutput"
}

func (AddOutput) ShortHelp() string {
	return "AddOutput <k> <n> <amount> -- Adds an output to a transaction.\n" +
		"                              the key <k> should be created by NewTransaction.\n" +
		"                              The address or name <n> can come from your address\n" +
		"                              book."
}

func (AddOutput) LongHelp() string {
	return `
AddOutput <key> <n|a> <amt>         <key>  created by a previous NewTransaction call
                                    <n|a>  A Valid Name in your Factoid Address 
                                           book, or a valid Factoid Address 
                                    <amt>  to be used to purchase Entry Credits at the
                                           current exchange rate.
`
}

/************************************************************
 * AddECOutput
 ************************************************************/
type AddECOutput struct {
	ICommand
}

// AddECOutput <key> <name|address> amount
//
// Buy Entry Credits

func (AddECOutput) Execute(state IState, args []string) error {

	if len(args) != 4 {
		return fmt.Errorf("Invalid Parameters")
	}
	key := args[1]
	adr := args[2]
	amt := args[3]

	ib := state.GetFS().GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(key))
	trans, ok := ib.(fct.ITransaction)
	if ib == nil || !ok {
		return fmt.Errorf("Unknown Transaction")
	}

	var addr fct.IAddress
	if !fct.ValidateECUserStr(adr) {
		if len(adr) != 64 {
			if len(adr) > 32 {
				return fmt.Errorf("Invalid Name.  Name is too long: %v characters", len(adr))
			}

			we := state.GetFS().GetDB().GetRaw([]byte(fct.W_NAME), []byte(adr))

			if we != nil {
				we2 := we.(wallet.IWalletEntry)
				addr, _ = we2.GetAddress()
				adr = hex.EncodeToString(addr.Bytes())
			} else {
				return fmt.Errorf("Name is undefined.")
			}
		} else {
			if badHexChar.FindStringIndex(adr) != nil {
				return fmt.Errorf("Invalid Name.  Name is too long: %v characters", len(adr))
			}
		}
	} else {
		addr = fct.NewAddress(fct.ConvertUserStrToAddress(adr))
	}
	amount, _ := fct.ConvertFixedPoint(amt)
	bamount, _ := strconv.ParseInt(amount, 10, 64)
	err := state.GetFS().GetWallet().AddECOutput(trans, addr, uint64(bamount))
	if err != nil {
		return err
	}

	fmt.Println("Added Output of ", amt, " to be paid to ", args[2],
		fct.ConvertECAddressToUserStr(addr))

	return nil
}

func (AddECOutput) Name() string {
	return "AddECOutput"
}

func (AddECOutput) ShortHelp() string {
	return "AddECOutput <k> <n> <amount> -- Adds an Entry Credit output (ecoutput)to a \n" +
		"                              transaction <k>.  The Entry Credits are assigned to\n" +
		"                              the address <n>.  The output <amount> is specified in\n" +
		"                              factoids, and purchases Entry Credits according to\n" +
		"                              the current exchange rage."
}

func (AddECOutput) LongHelp() string {
	return `
AddECOutput <key> <n|a> <amt>       <key>  created by a previous NewTransaction call
                                    <n|a>  Name or Address to hold the Entry Credits
                                    <amt>  Amount of Factoids to be used in this purchase.  Note
                                           that the exchange rate between Factoids and Entry
                                           Credits varies.
`
}

/************************************************************
 * Sign
 ************************************************************/
type Sign struct {
	ICommand
}

// Sign <k>
//
// Sign the given transaction identified by the given key
func (Sign) Execute(state IState, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("Invalid Parameters")
	}
	key := args[1]
	// Get the transaction
	ib := state.GetFS().GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(key))
	trans, ok := ib.(fct.ITransaction)
	if !ok {
		return fmt.Errorf("Invalid Parameters")
	}

	err := state.GetFS().GetWallet().Validate(1, trans)
	if err != nil {
		return err
	}

	ok, err = state.GetFS().GetWallet().SignInputs(trans)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("Error signing the transaction")
	}

	// Update our map with our new transaction to the same key.  Otherwise, all
	// of our work will go away!
	state.GetFS().GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), trans)

	return nil

}

func (Sign) Name() string {
	return "Sign"
}

func (Sign) ShortHelp() string {
	return "Sign <k> -- Sign the transaction given by the key <k>"
}

func (Sign) LongHelp() string {
	return `
Sign <key>                          Signs the transaction specified by the given key.
                                    Each input is found within the wallet, and if 
                                    we have the private key for that input, we 
                                    sign for that input.  
                                    
                                    Transctions can have inputs from multiple parties.
                                    In this case, the inputs can be signed by each
                                    party by first creating all the inputs and 
                                    outputs for a transaction.  Then signing your
                                    inputs.  Exporting the transaction.  Then
                                    sending it to the other party or parties for
                                    their signatures.
`
}

/************************************************************
 * Submit
 ************************************************************/
type Submit struct {
	ICommand
}

// Submit <k>
//
// Submit the given transaction identified by the given key
func (Submit) Execute(state IState, args []string) error {

	if len(args) != 2 {
		return fmt.Errorf("Invalid Parameters")
	}
	key := args[1]
	// Get the transaction
	ib := state.GetFS().GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(key))
	trans, ok := ib.(fct.ITransaction)
	if !ok {
		return fmt.Errorf("Invalid Parameters")
	}

	err := state.GetFS().GetWallet().Validate(1, trans)
	if err != nil {
		return err
	}

	err = state.GetFS().GetWallet().ValidateSignatures(trans)
	if err != nil {
		return err
	}

	// Okay, transaction is good, so marshal and send to factomd!
	data, err := trans.MarshalBinary()
	if err != nil {
		return err
	}

	transdata := string(hex.EncodeToString(data))

	s := struct{ Transaction string }{transdata}

	j, err := json.Marshal(s)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf("http://%s/v1/factoid-submit/", state.GetServer()),
		"application/json",
		bytes.NewBuffer(j))

	if err != nil {
		return fmt.Errorf("Error coming back from server ")
	}
	resp.Body.Close()

	// Clear out the transaction
	state.GetFS().GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), nil)

	fmt.Println("Transaction", key, "Submitted")

	return nil
}

func (Submit) Name() string {
	return "Submit"
}

func (Submit) ShortHelp() string {
	return "Submit <k> -- Submit the transaction given by the key <k>"
}

func (Submit) LongHelp() string {
	return `
Submit <key>                        Submits the transaction specified by the given key.
                                    Each input in the transaction must have  a valid
                                    signature, or Submit will reject the transaction.
`
}


