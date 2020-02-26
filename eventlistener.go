package main

import (
	"fmt"
	"os"
	"time"
	"math"
	"context"
	"io/ioutil"
	"encoding/json"
	"google.golang.org/grpc"
	"github.com/pkg/errors"
	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/common/crypto"
	"github.com/hyperledger/fabric/common/localmsp"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/orderer"
	common2 "github.com/hyperledger/fabric/peer/common"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/hyperledger/fabric/core/comm"
)

type DeliveryClient interface {
	Send(*common.Envelope) error
	Recv() (*peer.DeliverResponse, error)
}

type GRPCClient struct {
	grpcClient 		*comm.GRPCClient
	grpcClientConn		*grpc.ClientConn
	deliveryClient 		DeliveryClient
	signer      		crypto.LocalSigner
	tlsCertHash 		[]byte
}

var (
	fabric_cfg_path 	string
	msp_id 			string
	msp_type 		string
	msp_config_dir 		string
	client_key 		string
	client_cert 		string
	root_cert 		string
	server 			string
	channel_id 		string
	config_file		string	
)

func InitClientConfigs() (comm.ClientConfig,error) {

	err := ReadConfigs()
	if err != nil {
		return comm.ClientConfig{}, errors.WithMessage(err, "failed to read config json")
	}

	/* Initialize Config */
	err = common2.InitConfig(config_file)
	if err != nil { 
		return comm.ClientConfig{}, errors.WithMessage(err, "fatal error when initializing config")
	}

	/* Initialize Crypto */
	err = common2.InitCrypto(msp_config_dir, msp_id, msp_type)
	if err != nil { 
		return comm.ClientConfig{}, errors.WithMessage(err, "Cannot run client because")
	}

	/* Init Client Configs */
	clientConfig := comm.ClientConfig{
		KaOpts:  comm.DefaultKeepaliveOptions,
		SecOpts: &comm.SecureOptions{},
		Timeout: 5 * time.Minute,
	}

	clientConfig.SecOpts.UseTLS = true
	clientConfig.SecOpts.RequireClientCert = true
	
	rootCert, err := ioutil.ReadFile(root_cert)
	if err != nil {
		return comm.ClientConfig{}, errors.WithMessage(err, "Error loading TLS root certificate")
	}
	clientConfig.SecOpts.ServerRootCAs = [][]byte{rootCert}

	
	clientKey, err := ioutil.ReadFile(client_key)
	if err != nil {
		return comm.ClientConfig{}, errors.WithMessage(err, "Error loading client TLS key")
	}
	clientConfig.SecOpts.Key = clientKey

	clientCert, err := ioutil.ReadFile(client_cert)
	if err != nil {
		return comm.ClientConfig{}, errors.WithMessage(err, "Error loading client TLS cert")
	}
	clientConfig.SecOpts.Certificate = clientCert


	fmt.Println(" Crypto & Client Configs initialized Successfully!")

	return clientConfig, nil
}


func InitGRPCClient(clientConfig comm.ClientConfig) (*GRPCClient, error){

	grpcClient, err := comm.NewGRPCClient(clientConfig)
	if err != nil {
		fmt.Println("Error creating grpc client: "+ err.Error())
		return nil, errors.WithMessage(err, "Error creating grpc client")
	}

	grpcClientConn, err := grpcClient.NewConnection(server, "")
	if err != nil {
		fmt.Println("Error connecting: "+ err.Error())
		return nil, errors.WithMessage(err, "Error connecting")
	}

	signer := localmsp.NewSigner()
	tlsCertHash := util.ComputeSHA256(grpcClient.Certificate().Certificate[0])	

	fmt.Println(" GRPC Client initialized Successfully!")

	return &GRPCClient {
		grpcClient: grpcClient,
		grpcClientConn: grpcClientConn,
		signer: signer,
		tlsCertHash: tlsCertHash,
	}, nil
}


func(grpc *GRPCClient) InitDeliveryClient() (error){

	deliverClient := peer.NewDeliverClient(grpc.grpcClientConn)
	if deliverClient == nil {
		return errors.New("No Host Available")
	}

	var err error
	var deliveryClient DeliveryClient
	
	deliveryClient, err = deliverClient.Deliver(context.Background())
	if err != nil {
		return errors.WithMessage(err, "failed to connect")
	}

	grpc.deliveryClient = deliveryClient

	fmt.Println(" GRPC Delivery Client initialized Successfully!")

	return nil
}

func(grpc *GRPCClient) CreateEventStream() error{

	envelope, err := grpc.CreateSignedEnvelope() 
	if err != nil {
		return errors.WithMessage(err, "Error creating signed envelope")
	}

	err = grpc.deliveryClient.Send(envelope)
	if err != nil {
		return errors.WithMessage(err, "Error in delivering the signed envelope")
	}
	fmt.Println(" GRPC Event Stream created Successfully!")
	return nil
}

func(grpc *GRPCClient) CreateSignedEnvelope() (*common.Envelope,error) {

	start  := &orderer.SeekPosition{
		Type: &orderer.SeekPosition_Newest{
			Newest: &orderer.SeekNewest{},
		},
	}

	stop := &orderer.SeekPosition{
		Type: &orderer.SeekPosition_Specified{
			Specified: &orderer.SeekSpecified{
				Number: math.MaxUint64,
			},
		},
	}

	env, err := utils.CreateSignedEnvelopeWithTLSBinding(common.HeaderType_DELIVER_SEEK_INFO, 
		channel_id, grpc.signer, 
		&orderer.SeekInfo{
			Start:    start,
			Stop:     stop,
			Behavior: orderer.SeekInfo_BLOCK_UNTIL_READY,
	}, 0, 0, grpc.tlsCertHash)
	if err != nil {
		return nil, errors.WithMessage(err, "Error creating signed envelope")
	}

	fmt.Println(" Signed Envelope Created ")
	fmt.Println(" Seek Info Start = ", start)
	fmt.Println(" Seek Info Stop = ", stop)

	return env, nil
}


func(grpc *GRPCClient) ReadEventStream() error{

	fmt.Println(" \n Started listening GRPC Event Stream at ", server)

	for {

		receivedMsg, err := grpc.deliveryClient.Recv()
		if err != nil {
			return errors.WithMessage(err, "Error in Receiving Message")
		} 

		switch t := receivedMsg.Type.(type) {

			case *peer.DeliverResponse_Status:
				fmt.Println(" Received DeliverResponse Status = ", t)
			case *peer.DeliverResponse_Block:				
				fmt.Println(" Received a new Block from = ",channel_id)
				ReadBlock(t.Block)			
				
		}	
	}

	return nil
}



func ReadConfigs() error {

	ROOT := os.Getenv("GOPATH")
	
	configFile, err := os.Open("config.json")
	if err != nil {
		return errors.WithMessage(err, "failed to read config json")
	}
	defer configFile.Close()


	byteValue, _ := ioutil.ReadAll(configFile)
	var configs map[string]interface{}
	json.Unmarshal([]byte(byteValue), &configs)

	fabric_cfg_path = ROOT + fmt.Sprint(configs["fabric_cfg_path"])
	msp_id = fmt.Sprint(configs["msp_id"])
	msp_type = fmt.Sprint(configs["msp_type"])
	msp_config_dir = fabric_cfg_path + fmt.Sprint(configs["msp_config_dir"])
	client_key = fabric_cfg_path + fmt.Sprint(configs["client_key"])
	client_cert = fabric_cfg_path + fmt.Sprint(configs["client_cert"])
	root_cert = fabric_cfg_path + fmt.Sprint(configs["root_cert"])
	server = fmt.Sprint(configs["server"])
	channel_id = fmt.Sprint(configs["channel_id"])
	config_file = fmt.Sprint(configs["config_file"])

	fmt.Println(" ### Configs ")
	fmt.Println(" ROOT = ", ROOT)
	fmt.Println(" FABRIC_CFG_PATH = ",configs["fabric_cfg_path"])
	fmt.Println(" MSP ID = ",configs["msp_id"])
	fmt.Println(" MSP TYPE = ",configs["msp_type"])
	fmt.Println(" MSP CONFIG DIR = ",configs["msp_config_dir"])
	fmt.Println(" CLIENT KEY = ",configs["client_key"])
	fmt.Println(" CLIENT CERT = ",configs["client_cert"])
	fmt.Println(" ROOT CERT = ",configs["root_cert"])
	fmt.Println(" GRPC LISTENING SERVER = ",configs["server"])
	fmt.Println(" CHANNEL ID = ",configs["channel_id"])
	fmt.Println(" CONFIG FILE = ",configs["config_file"])

	return nil
}
