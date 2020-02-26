package main

import (
	"fmt"
)

func main(){

	clientConfig, err := InitClientConfigs()
	if err != nil {
		fmt.Println("Error when initializing Client Config "+ err.Error())
	}

	fmt.Println("Client Configs initialized successfully")

	grpcClient, err := InitGRPCClient(clientConfig)
	if err != nil {
		fmt.Println(" Error when initializing GRPC Client Connection "+err.Error())
	}

	err = grpcClient.InitDeliveryClient(false)
	if err != nil {
		fmt.Println(" Error when intializing Delivery Client "+err.Error())
	}

	err = grpcClient.CreateEventStream()
	if err != nil {
		fmt.Println(" Error when Creating Event Stream - "+err.Error())
	}

	err = grpcClient.ReadEventStream()
	if err != nil {
		fmt.Println(" Error reading event stream - "+err.Error())
	}

}