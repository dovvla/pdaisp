# pdaisp

End of course project for Parallel and Distributed Architectures and Data Structures

In order to run project install prerequisites for hyperledger-fabric 2.2.6, then position to `test-network` directory from sample and execute the following command:
`./network.sh down; ./network.sh up createChannel -ca; cd addOrg4/; ./addOrg4.sh up -ca; cd ..; cd addOrg3/; ./addOrg3.sh up -ca; cd ..;./network.sh deployCC -ccn basic -ccp ../go-chaincodes/ -ccl go;`

Client code is available in `go-client` directory and can be run with `go run main.go`
