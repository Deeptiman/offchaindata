package main

import (
	"fmt"
	"context"
	"encoding/json"
	kivik "github.com/go-kivik/kivik/v4"
	_ "github.com/go-kivik/couchdb/v4"
	"github.com/pkg/errors"
)

func SaveToCouchDB(kvWriteSet []byte) error{

	fmt.Println("Saving to CouchDB")
	client, err := kivik.New("couch", "http://localhost:5990/")	

	if err != nil {
		fmt.Println("failed to set couchdb - "+err.Error())
		return errors.WithMessage(err,"failed to set couchdb: ")
	}

	db := client.DB(context.TODO(), "offchaindb")

	/*if err != nil {
		return errors.WithMessage(err,"failed to create database: ")
	}*/

	User := &SampleUser{}
	err = json.Unmarshal(kvWriteSet, User)
	if err != nil{
		return errors.WithMessage(err,"unmarshaling write set error: ")
	}

	rev, err := db.Put(context.TODO(), User.Email, User)
	if err != nil {
		fmt.Println("user insertion error - "+err.Error())
		return errors.WithMessage(err,"user insertion error: ")
	}

	fmt.Println("User Inserted into CouchDB %s ",rev)

	return nil
}