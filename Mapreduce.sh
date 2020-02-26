## Create MapReduce for Email
curl -X PUT http://127.0.0.1:5990/offchaindb/_design/emailviewdesign/ -d '{"views":{"emailview":{"map":"function(doc) { emit(doc.email,1);}", "reduce":"function (keys, values, combine) {return sum(values)}"}}}' -H 'Content-Type:application/json'

## Query Reduce
curl -X GET http://127.0.0.1:5990/offchaindb/_design/emailviewdesign/_view/emailview?reduce=true

## Query Group
curl -X GET http://127.0.0.1:5990/offchaindb/_design/emailviewdesign/_view/emailview?group=true

