# Compound Connector

This is an example compound connector that uses `EthereumConnector` as a base connector.

# How to Build a Connector

## Prerequisites

1. Golang 1.20+
2. Docker
3. protoc
4. protoc-gen-go
5. abigen
6. solc

## Overview

We will be building a source connector. As a reminder,a source connector extracts/ingests, cleans, and generalizes data
from an external source, usually a blockchain, smart contract, or API. It then passes on this data to the Nakji message
queue. **It is highly recommended to use the example compound connector as the template to build your own connector.**

### Code of a Connector

#### 1. Getting the ABI files for a smart contract

ABI stands for Application Binary Interface. It is a file which exists for each smart contract, and enables it to
communicate and interact with external applications and other smart contracts. You can read more about what ABI files do
here.

To find the ABI file associated with a given smart contract, I recommend looking up the contract address for that smart
contract on Google, or checking the docs of the organization responsible for maintaining the contract. Next, you can
paste that address into a blockchain explorer like Etherscan.io, which looks up the contract according to its address
and displays useful information such as the ABI file, contract source code, and live events handled by the contract. We
will use the latter to verify that our connector is listening to the events correctly.

Let’s use the compound connector as an example. Fortunately, I found all abi in a single
file [mainnet-abi.json](smart-contracts/mainnet-abi.json), but unfortunately the entire file is not a valid format, you
need to copy the parts you need to somewhere else and generate `abi.go` files.

#### 2. Creating the connector directory

The folder containing your connector should be organized like this. You should create the folders in your connector
directory, using the <connector_name> and <contract_name>. For convention, the <connector_name> should be all lower
case, the <contract_name> all upper case, with no spaces or punctuation.

```
<connector_name>/
    cmd/<connector_name>/
        main.go
    <contract_name>/
        <contract_name>.abi
        <contract_name>.abi.go
        <contract_name>.proto
        <contract_name>.pb.go
        <contract_name>.go
    addresses.go
    contract.go
    local.yaml
    manifest.yaml
    types.go
    <connector_name>.go
    Additional files 
```

If you are new to Golang, go code is organized in directories called packages–every `.go` file contains a package
declaration line at the head of the file. For your connector, each `.go` file in <contract_name>/ should be declared as
part of <contract_name> package. Each `.go` file in <connector_name>/ should be declared as part of <connector_name>
package. The `main.go` file is part of the main package, the Go executable designation.

#### 3. Creating the contract handling files

First, enter the <contract_name>/ directory. Create the <contract_name>.abi file in <contract_name>/ and paste the ABI
source code you copied earlier.

Next, generate the `<contract_name>.abi.go` file using `abigen` (you should have `abigen` installed from the above step)
.

**Need to add abigen and protogen installation guide links here**

```shell
abigen --abi <contract_name>.abi --pkg <contract_name> --out <contract_name>.abi.go
```

Next, generate the <contract_name>.proto file using `protogen`(you should have `protogen` installed from the above step)
.

```shell
protogen -p <contract_name> -i "./<contract_name>.abi" <contract_name>.proto
```

This creates a `.proto` file, which defines the data format for a protobuf message. A `.proto` file can define multiple
messages; in our case each message corresponds to a type of event that the smart contract can emit. It should look
something like this.

```protobuf
syntax = "proto3";
package cBAT;

import "google/protobuf/timestamp.proto";

option go_package = "github.com/nakji-network/connector/examples/compound/cBAT";

// Mint represents a Mint event raised by the Compound contract.
message Mint {
  google.protobuf.Timestamp ts = 1;
  uint64 block = 2;
  uint64 idx = 3;
  bytes tx = 4; // tx hash
  bytes minter = 5; // The address that minted the assets
  bytes mintAmount = 6;
  bytes mintTokens = 7;
}

// Redeem represents a Borrow event raised by the Compound contract.
message Redeem {
  google.protobuf.Timestamp ts = 1;
  uint64 block = 2;
  uint64 idx = 3;
  bytes tx = 4; // tx hash
  bytes redeemer = 5; // The address that redeemed the assets
  bytes redeemAmount = 6;
  bytes redeemTokens = 7;
}

// Borrow represents a Borrow event raised by the Compound contract.
message Borrow {
  google.protobuf.Timestamp ts = 1;
  uint64 block = 2;
  uint64 idx = 3;
  bytes tx = 4; // tx hash
  bytes borrower = 5; // The address that borrowed the assets
  bytes borrowAmount = 6;
  bytes accountBorrows = 7;
  bytes totalBorrows = 8;
}
```

Since the `.proto` file is generated from the ABI file, it only contains the timestamp and event-related fields for each
message type. For our data indexing purposes, we need to add some additional fields which will serve as unique
identifiers when receiving events. These are values describing the event’s location on the chain, and are logged by the
eth client. Which values to include is a connector-specific design decision you should discuss with others.

Next, generate the `<contract_name>.pb.go` file using `protoc`.

```shell
protoc --go_out=. --go_opt=paths=source_relative ./<contract_name>.proto
```

Lastly, we will need to create a <contract_name>.go file (no generator this time, unfortunately). It defines the
SmartContract struct and implements the `*ProtoMessageGetter.Get(...)` function. You can copy the code structure from
[cBAT.go](cBAT/cBAT.go) as an example.

```go
package cBAT

import (
...
)
type EventParser struct{}

func (ep *EventParser) Get(eventName string, contractAbi *abi.ABI, evLog types.Log, timestamp *timestamppb.Timestamp) proto.Message {
	switch eventName {
	case "Mint":
		event := new(CBATMint)
		if err := common.UnpackLog(*contractAbi, event, eventName, evLog); err != nil {
			log.Error().Err(err).Msg("Unpack Mint event error")
			return nil
		}
		return &Mint{
			Ts:         timestamp,
			Block:      evLog.BlockNumber,
			Idx:        uint64(evLog.Index),
			Tx:         evLog.TxHash.Bytes(),
			Minter:     event.Minter.Bytes(),
			MintAmount: event.MintAmount.Bytes(),
			MintTokens: event.MintTokens.Bytes(),
		}
	}
	return nil
}
```

First, you need to change the case values inside switch eventName {...} to the names of possible events that your smart
contract can handle. The list of event names can be found in the struct definitions in `<contract_name>.pb.go`. They
also
correspond to message types in `<contract_name>.proto`.

Second, you need to replace `Mint` in `event := new(Mint)` with the struct definitions in `<contract_name>.abi.go`.
Usually, these are in the format `"contract_name"+"event_name"`, but check to make sure it matches.

Third, you need to modify the function to return the address of a specific event struct, so you will also need to pass
in values for timestamp, event-specific fields, and any other fields that you added to the `.proto` file. For
event-specific fields, you can retrieve the value from `evLog`, converting it to bytes first. For additional fields, you
retrieve them from `evLog` (log object handled by the eth client). Check the definition of `evLog` using your IDE and
make
sure the original data type matches the field type you defined in the `.proto` file (i.e. `evLog.Index` is a `uint`, so
you’ll
need to cast it into `uint64`).

Do this for each event type. Essentially, `<contract_name>.go` converts the event data returned by the eth client (
defined
in `.abi.go`) into protobuf format, mapping all data to their corresponding fields. You are now done with the
contract-handling part of the connector!

#### 4. Create the connector runtime files

Now you will create the files to allow your connector to call the Nakji connector library and use it to instantiate and
run your connector. These include all the files in the <connector_name>/ directory. You can start by copying the files
from `compound/` into `<connector_name>/`. Then you need to modify each file to fit your connector.

* `contract.go`
* `local.yaml`
* `manifest.yaml`
* `types.go`
* `<connector_name>.go`

`contract.go`: This defines the functions to retrieve the Ethereum ABI from the contract address and save it to a field
in
our connector struct. You can copy this file directly from `compound`.

`local.yaml`: This is the configuration file containing the rpc urls which are needed to retrieve data from the
blockchain. You’ll need it to run and test your connector locally.

`manifest.yaml`: This contains info about the connector name, author, version etc… Name should be <
connector_name>, author is `nakji`, and the version is `0.0.0`.

`types.go`: This defines the maps used by main.go. TopicTypes maps from string to protobuf format structs defined in <
contract_name>.pb.go. The string keys should be all lowercase, with the format `<author>.<connector_name>.<version>.<
contract_name>_<event_name>`. The struct values should match the names of structs in `<contract_name>.pb.go`. ABIs maps
from <contract_name> to the ABI variable in `<contract_name>.abi.go`.

`<connector_name>.go`: This defines the start and parse functions that the connector uses during runtime to start
receiving and unpacking messages. You can copy the code from `compound.go`, changing the contract name to your specific
contract.

#### 5. Creating the main executable

Now we just need to create the `main.go` file in `<connector_name>/cmd/<connector_name>/`. This is the main executable
which
handles starting and running the connector, as well as receiving user flags for backfill. Copy the code from
`compound/cmd/main.go`, changing the contract name. You also need to change the network name in the config to the
chain your contract is deployed on. Congrats! You just built your first connector.

#### 6. Running a connector locally

You should now be able to run your connector locally without any additional steps. From your terminal, go into <
connector_name>/ and run

```shell
go mod init

go mod tidy
```

This automatically creates the `go.mod` and `go.sum` files that tell go to fetch which modules and dependencies your
program
depends on.

Next, spin up kafka by running:

```shell
docker-compose up -d
```

Now start the connector locally by running the `main.go` file. (If you are running to “connection refused” error, make
sure docker is running)

```shell
go run cmd/<connector_name>/main.go
```

If your connector is functioning correctly, you should see the terminal continuously output some log like header
received block=20889028 network=ethereum ts=1661846917 , indicating that the connector is successfully receiving the
current
block number.

Next, check that your connector is receiving live events. Leave your connector running, go back to the webpage on the
blockchain explorer for your smart contract, and navigate to the "events" tab (next to the "contracts" tab). It lists
live events as they come in, so refresh the page every once in a while.

Hover over the timestamp for a new event in the webpage. In the terminal where your connector is running, you should see
a log like “Delivered message key= offset=257 partition=0 topic=dev.fct.nakji.compound.0_0_0.cBAT_Mint” followed by
"successfully committed transactions". Check that the timestamp of this log matches the timestamp for the event in the
webpage. This means that your connector has successfully received the event, and committed the data as a message into
Kafka.

Leave your connector running for a while. If the logs outputted by your connector all match the new events on the online
explorer, then your connector is working as intended. 
