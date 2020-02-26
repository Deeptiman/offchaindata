package main

import (
	"fmt"
	"encoding/json"
	//"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/ledger/rwset/kvrwset"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/ledger/rwset"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

func ReadBlock(block *common.Block) error{

	blockData := block.Data.Data

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

	fmt.Println("Chaincode Name === "+input.ChaincodeSpec.ChaincodeId.Name)

	chaincodeArgs := make([]string, len(input.ChaincodeSpec.Input.Args))

	for i, c := range input.ChaincodeSpec.Input.Args {
		args := CToGoString(c[:])
		chaincodeArgs[i] = args
	}

	fmt.Println("Chaicode Args = ", chaincodeArgs)

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

	if len(kvrwset.Reads) != 0 {
		
		fmt.Println("BlockNum = ",kvrwset.Reads[0].Version.BlockNum)
		fmt.Println("TxNum = ",kvrwset.Reads[0].Version.TxNum)
		fmt.Println("KVRead Key = ",kvrwset.Reads[0].Key)
		//fmt.Println("KVRead Value = ",kvrwset.Reads[0].Value)
	}

	if len(kvrwset.Writes) != 0 {

		args := CToGoString(kvrwset.Writes[0].Value[:])

		fmt.Println("KVWrite Key = ",kvrwset.Writes[0].Key)
		fmt.Println("KVWrite Value = ",args)

		User := &SampleUser{}
		err = json.Unmarshal(kvrwset.Writes[0].Value, User)
		if err != nil{
			return errors.WithMessage(err,"unmarshaling write set error: ")
		}

		fmt.Println("Email  = "+User.Email)
		fmt.Println("Name = "+User.Name)
		fmt.Println("Age = "+User.Age)
		fmt.Println("Country = "+User.Country)

		SaveToCouchDB(kvrwset.Writes[0].Value)
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