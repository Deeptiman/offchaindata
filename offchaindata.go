package main

import (
	"fmt"
)

func main(){

	fmt.Println("************************************************")
	fmt.Println(" Hyperledger Fabric OffChain Storage Demo ")
	fmt.Println("************************************************")

	
	clientConfig, err := InitClientConfigs()
	if err != nil {
		fmt.Println("Error when initializing Client Config "+ err.Error())
		return
	}


	grpcClient, err := InitGRPCClient(clientConfig)
	if err != nil {
		fmt.Println(" Error when initializing GRPC Client Connection "+err.Error())
		return
	}

	err = grpcClient.InitDeliveryClient()
	if err != nil {
		fmt.Println(" Error when intializing Delivery Client "+err.Error())
		return
	}

	err = grpcClient.CreateEventStream()
	if err != nil {
		fmt.Println(" Error when Creating Event Stream - "+err.Error())
		return
	}

	err = grpcClient.ReadEventStream()
	if err != nil {
		fmt.Println(" Error reading event stream - "+err.Error()) 
	}

}