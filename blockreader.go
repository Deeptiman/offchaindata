package main

import (
	"fmt"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/ledger/rwset/kvrwset"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/ledger/rwset"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

func ReadBlock(block *common.Block) error{

	fmt.Println("\n ######################################################################## ")

	blockHeader := block.Header
	blockData := block.Data.Data

	fmt.Println(" Received Block Number = ", blockHeader.Number)

	//First Get the Envelope from the BlockData
	envelope, err := GetEnvelopeFromBlock(blockData[0])
	if err != nil {
		return errors.WithMessage(err,"unmarshaling Envelope error: ")
	}

	//Retrieve the Payload from the Envelope
	payload := &common.Payload{}
	err = proto.Unmarshal(envelope.Payload, payload)
	if err != nil {
		return errors.WithMessage(err,"unmarshaling Payload error: ")
	}

	transaction := &peer.Transaction{}
	err = proto.Unmarshal(payload.Data, transaction)
	if err != nil {
		return errors.WithMessage(err,"unmarshaling Payload Transaction error: ")
	}

	chaincodeActionPayload := &peer.ChaincodeActionPayload{}
	err = proto.Unmarshal(transaction.Actions[0].Payload, chaincodeActionPayload)
	if err != nil {
		return errors.WithMessage(err,"unmarshaling Chaincode Action Payload error: ")
	}

	chaincodeProposalPayload := &peer.ChaincodeProposalPayload{}
	err = proto.Unmarshal(chaincodeActionPayload.ChaincodeProposalPayload, chaincodeProposalPayload)
	if err != nil {
		return  errors.WithMessage(err,"unmarshaling Chaincode Proposal Payload error: ")
	}
 
	// The Input field is marshalled object of ChaincodeInvocationSpec
	input := &peer.ChaincodeInvocationSpec{}
	err = proto.Unmarshal(chaincodeProposalPayload.Input, input)
	if err != nil {
		return  errors.WithMessage(err,"unmarshaling Chaincode Proposal Payload Input error: ")
	}

	chaincodeArgs := make([]string, len(input.ChaincodeSpec.Input.Args))
	for i, c := range input.ChaincodeSpec.Input.Args {
		args := CToGoString(c[:])
		chaincodeArgs[i] = args
	}

	fmt.Println("\n ## Chaincode ")
	fmt.Println(" Name : ", input.ChaincodeSpec.ChaincodeId.Name)
	fmt.Println(" Version : ", input.ChaincodeSpec.ChaincodeId.Version)
	fmt.Println(" Args : ", chaincodeArgs)


	proposalResponsePayload	:= &peer.ProposalResponsePayload{}
	err = proto.Unmarshal(chaincodeActionPayload.Action.ProposalResponsePayload, proposalResponsePayload)
	if err != nil {
		return errors.WithMessage(err,"unmarshaling Proposal Response Payload error: ")
	}

	chaincodeAction := &peer.ChaincodeAction{}
	err = proto.Unmarshal(proposalResponsePayload.Extension, chaincodeAction)
	if err != nil {
		return errors.WithMessage(err,"unmarshaling Extension error: ")
	}

	txReadWriteSet := &rwset.TxReadWriteSet{}
	err = proto.Unmarshal(chaincodeAction.Results, txReadWriteSet)
	if err != nil {
		return errors.WithMessage(err,"unmarshaling txReadWriteSet error: ")
	}

	RwSet := txReadWriteSet.NsRwset[0].Rwset

	kvrwset := &kvrwset.KVRWSet{}
	err = proto.Unmarshal(RwSet, kvrwset)
	if err != nil {
		return errors.WithMessage(err,"unmarshaling kvrwset error: ")
	}

	fmt.Println("\n Block Read Write Set ")
	if len(kvrwset.Reads) != 0 {		
		fmt.Println("\n ## KVRead Set ")
		fmt.Println(" BlockNum = ",kvrwset.Reads[0].Version.BlockNum)
		fmt.Println(" TxNum = ",kvrwset.Reads[0].Version.TxNum)
		fmt.Println(" Key = ",kvrwset.Reads[0].Key)
	}

	if len(kvrwset.Writes) != 0 {

		values := CToGoString(kvrwset.Writes[0].Value[:])

		fmt.Println("\n ## KVWrite Set")
		fmt.Println(" Key = ",kvrwset.Writes[0].Key)
		fmt.Println(" Value = ",values)

		err = SaveToCouchDB(kvrwset.Writes[0].Value)
		if err != nil {
			return errors.WithMessage(err,"error while saving to CouchDB")
		}
	}



	return nil

}

func GetEnvelopeFromBlock(data []byte) (*common.Envelope, error){

	var err error
	env := &common.Envelope{}
	if err = proto.Unmarshal(data, env); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling Envelope")
	}

	return env, nil
}

func CToGoString(c []byte) string {
    n := -1
    for i, b := range c {
        if b == 0 {
            break
        }
        n = i
    }
    return string(c[:n+1])
}