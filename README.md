<h1>OffChain Data</h1>
<p>     <img alt="GitHub last commit" src="https://img.shields.io/github/last-commit/Deeptiman/offchaindata">  <img alt="GitHub language count" src="https://img.shields.io/github/languages/count/Deeptiman/offchaindata"> <img alt="GitHub top language" src="https://img.shields.io/github/languages/top/Deeptiman/offchaindata"></p>
<p><a href="https://www.hyperledger.org/projects/fabric"><img src="https://www.hyperledger.org/wp-content/uploads/2016/09/logo_hl_new.png" alt="N|Solid"></a></p>
<p><b>OffChain Data</b> is a sample demonstration to understand the concept of implementing offchain storage and it's capability in Hyperledger fabric Blockchain network.
 So, this project will work as a peer block event listener and will store the block details in the <b>CouchDB</b> that be query through <b>MapReduce</b>.</p>
 <p><b>Medium writeup : </b><a href="https://medium.com/@deeptiman/offchain-storage-in-hyperledger-fabric-77e28bd99e0c">https://medium.com/@deeptiman/offchain-storage-in-hyperledger-fabric-77e28bd99e0c</a>

 
 <h2>Configuration requirements</h2>
 <p>You need to add the certain project details in `config.json`, so that it will be used to create an event listener and the Blocks will be received through <b>GRPC</b> delivery
client. </p>

````````````````````````````````````````````````````````````````````````````````````````````````````````````
export FABRIC_CFG_PATH= /home/user/go/src/github.com/exampleledger/fixtures
````````````````````````````````````````````````````````````````````````````````````````````````````````````

````````````````````````````````````````````````````````````````````````````````````````````````````````````
       {
            "peer_config_path": "exampleledger/fixtures/crypto-config/peerOrganizations/",
            "msp_id": "Org1MSP",
            "msp_type": "bccsp",    
            "msp_config_dir": "org1.example.ledger.com/users/Admin@org1.example.ledger.com/msp",
            "client_key": "org1.example.ledger.com/peers/peer0.org1.example.ledger.com/tls/server.key",
            "client_cert": "org1.example.ledger.com/peers/peer0.org1.example.ledger.com/tls/server.crt",
            "root_cert": "org1.example.ledger.com/peers/peer0.org1.example.ledger.com/tls/ca.crt",
            "server": "peer0.org1.example.ledger.com:7051",
            "channel_id": "exampleledger",
            "config_file": "configtx"
       }
`````````````````````````````````````````````````````````````````````````````````````````````````````````````````

<h2>Create CouchDB local instance </h2>

The CouchDB local instance can be created using Docker.
```````````````````````````````````````````````````````````````````````````````````````````````````````````````
docker run --publish 5990:5984 --detach --name offchaindb hyperledger/fabric-couchdb
docker start offchaindb
`````````````````````````````````````````````````````````````````````````````````````````````````````````````````

<h2> Mock Chaincode Model</h2>
I have followed a sample user model to create the offchaindb. You can also create your own chaincode model and the offchaindata
will listen the `KVWriteSet` to store in the couchdb.

<b>Sample Model</b>
``````````````````````````````````````````````````````````````````````````````````````````````````````````````````
type SampleUser struct {
	Email 	  string 		`json:"email"`	
	Name 	  string 		`json:"name"`
	Age	  string		`json:"age"`
	Country   string		`json:"country"`
 }
``````````````````````````````````````````````````````````````````````````````````````````````````````````````````

<h2>Configure MapReduce</h2>

<p><b>MapReduce</b> will query the offchain data from <b>CouchDB</b>. So, you need to configure MapReduce for certain design element from CouchDB collection.</p>

<b>Configure MapReduce for Email</b>
````````````````````````````````````````````````````````````````````````````````````````````````````````````````
curl -X PUT http://127.0.0.1:5990/offchaindb/_design/emailviewdesign/ -d '{"views":{"emailview":{"map":"function(doc) { emit(doc.email,1);}", "reduce":"function (keys, values, combine) {return sum(values)}"}}}' -H 'Content-Type:application/json'
````````````````````````````````````````````````````````````````````````````````````````````````````````````````
<b>Output</b>
````````````````````````````````````````````````````````````````````````````````````````````````````````````````
{"ok": true, "id":"_design/emailviewdesign", "rev": "1-f34147f686003ff5c7da5a5e7e2759b8"}
````````````````````````````````````````````````````````````````````````````````````````````````````````````````

<b>Query `Reduce` function to count total email</b>
```````````````````````````````````````````````````````````````````````````````````````````````````````````````
curl -X GET http://127.0.0.1:5990/offchaindb/_design/emailviewdesign/_view/emailview?reduce=true
```````````````````````````````````````````````````````````````````````````````````````````````````````````````
<b>Output</b> 
```````````````````````````````````````````````````````````````````````````````````````````````````````````````
{"rows":[
		{"key":null,"value":7}
	]}
```````````````````````````````````````````````````````````````````````````````````````````````````````````````

<b>Query `Map` function to list all emails</b>
```````````````````````````````````````````````````````````````````````````````````````````````````````````````
curl -X GET http://127.0.0.1:5990/offchaindb/_design/emailviewdesign/_view/emailview?group=true
```````````````````````````````````````````````````````````````````````````````````````````````````````````````
<b>Output</b>
```````````````````````````````````````````````````````````````````````````````````````````````````````````````
{"rows":[
		{"key":"alice@gmail.com","value":1},
		{"key":"john@gmail.com","value":1},
		{"key":"michale@gmail.com","value":1},
		{"key":"mark@mail.com","value":1},
		{"key":"bob@gmail.com","value":1},
		{"key":"oscar@gmail.com","value":1},
		{"key":"william@example.com","value":1}
	]}
```````````````````````````````````````````````````````````````````````````````````````````````````````````````

So, all the query peformed in offchain without querying from blockchain ledger.

<h2>License</h2>
<p>This project is licensed under the <a href="https://github.com/Deeptiman/offchaindata/blob/master/LICENSE">MIT License</a></p>
