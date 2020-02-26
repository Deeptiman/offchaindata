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
	//"github.com/hyperledger/fabric/common/tools/protolator"
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
	grpcClient 			*comm.GRPCClient
	grpcClientConn		*grpc.ClientConn
	deliveryClient 		DeliveryClient
	signer      		crypto.LocalSigner
	tlsCertHash 		[]byte
}

var (
	fabric_cfg_path 	string
	msp_id 				string
	msp_type 			string
	msp_config_dir 		string
	client_key 			string
	client_cert 		string
	root_cert 			string
	server 				string
	channel_id 			string
	config_file			string	
)

const (
	ROOT = "/home/harald/go/src/github.com/"
)

func InitClientConfigs() (comm.ClientConfig,error) {

	fmt.Println("Init Client Configs ")

	err := ReadConfigs()
	if err != nil {
		return comm.ClientConfig{}, errors.WithMessage(err, "failed to read config json")
	}

	/* Initialize Config */
	err = common2.InitConfig(config_file)
	if err != nil { 
		return comm.ClientConfig{}, errors.WithMessage(err, "fatal error when initializing config")
	}

	fmt.Println("Init Config Successfully")

	/* Initialize Crypto */
	err = common2.InitCrypto(msp_config_dir, msp_id, msp_type)
	if err != nil { 
		return comm.ClientConfig{}, errors.WithMessage(err, "Cannot run client because")
	}

	fmt.Println("Init Crypto Successfully")

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

	return &GRPCClient {
		grpcClient: grpcClient,
		grpcClientConn: grpcClientConn,
		signer: signer,
		tlsCertHash: tlsCertHash,
	}, nil
}


func(grpc *GRPCClient) InitDeliveryClient(filtered bool) (error){


	deliverClient := peer.NewDeliverClient(grpc.grpcClientConn)

	if deliverClient == nil {
		return errors.New("No Host Available")
	}

	var err error
	var deliveryClient DeliveryClient
	if filtered {
		deliveryClient, err = deliverClient.DeliverFiltered(context.Background()) 
	} else {
		deliveryClient, err = deliverClient.Deliver(context.Background())
	}

	if err != nil {
		return errors.WithMessage(err, "failed to connect")
	}

	grpc.deliveryClient = deliveryClient
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
	return env, nil
}


func(grpc *GRPCClient) ReadEventStream() error{

	for {

		receivedMsg, err := grpc.deliveryClient.Recv()
		if err != nil {
			return errors.WithMessage(err, "Error in Receiving Message")
		} 

		switch t := receivedMsg.Type.(type) {

			case *peer.DeliverResponse_Status:
				fmt.Println("Block Status ",t)
				//return nil
			case *peer.DeliverResponse_Block:
				
				fmt.Println(" Read Block --")
				ReadBlock(t.Block)
				
				/*err = protolator.DeepMarshalJSON(os.Stdout, t.Block)
				if err != nil {
					return errors.WithMessage(err, "  Error pretty printing block")
				}*/

				//return nil
			case *peer.DeliverResponse_FilteredBlock:
				
				fmt.Println(" Read Filter Block --")
				//ReadBlock(t.FilteredBlock)

				/*err = protolator.DeepMarshalJSON(os.Stdout, t.FilteredBlock)
				if err != nil {
					return errors.WithMessage(err, "  Error pretty printing filtered block")
				}*/
				//return nil
		}	
	}

	return nil
}



func ReadConfigs() error {

	fmt.Println("Read Configs ")

	configFile, err := os.Open("config.json")

	if err != nil {
		return errors.WithMessage(err, "failed to read config json")
	}

	fmt.Println("Successfully Opened Config File")

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

	fmt.Println("**** Configs **** ")
	fmt.Println(configs["fabric_cfg_path"])
	fmt.Println(configs["msp_id"])
	fmt.Println(configs["msp_type"])
	fmt.Println(msp_config_dir)
	fmt.Println(client_key)
	fmt.Println(client_cert)
	fmt.Println(root_cert)
	fmt.Println(configs["server"])
	fmt.Println(configs["channel_id"])
	fmt.Println(configs["config_file"])

	return nil
}
