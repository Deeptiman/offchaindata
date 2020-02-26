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

	dbName := "offchaindb"
	couchdbUrl := "http://localhost:5990/"

	fmt.Println("\n Saving the Block details to CouchDB")
	fmt.Println(" CouchDB URL = ", couchdbUrl)
	fmt.Println(" DB Name = ", dbName)

	client, err := kivik.New("couch", couchdbUrl)	
	if err != nil {
		fmt.Println("failed to set couchdb - "+err.Error())
		return errors.WithMessage(err,"failed to set couchdb: ")
	}

	db := client.DB(context.TODO(), dbName)

	User := &SampleUser{}
	err = json.Unmarshal(kvWriteSet, User)
	if err != nil{
		return errors.WithMessage(err,"unmarshaling write set error: ")
	}

	fmt.Println("\n Collection Key ", User.Email)

	rev, err := db.Put(context.TODO(), User.Email, User)
	if err != nil {
		fmt.Println(" Error during insertion - "+err.Error())
	} else {
		fmt.Println(" User Inserted into CouchDB - Revision = "+rev+" \n")
	}
	
	fmt.Println(" ######################################################################## \n")

	return nil
}