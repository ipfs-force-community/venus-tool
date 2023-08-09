<h1 align="center">Venus Tool</h1>

<p align="center">
 <a href="https://github.com/ipfs-force-community/venus-tool/actions"><img src="https://github.com/ipfs-force-community/venus-tool/actions/workflows/build_upload.yml/badge.svg"/></a>
 <a href="https://codecov.io/gh/ipfs-force-community/venus-tool"><img src="https://codecov.io/gh/ipfs-force-community/venus-tool/branch/master/graph/badge.svg?token=J5QWYWkgHT"/></a>
 <a href="https://goreportcard.com/report/github.com/ipfs-force-community/venus-tool"><img src="https://goreportcard.com/badge/github.com/ipfs-force-community/venus-tool"/></a>
 <a href="https://github.com/ipfs-force-community/venus-tool/tags"><img src="https://img.shields.io/github/v/tag/ipfs-force-community/venus-tool"/></a>
  <br>
</p>

## Why

We created this tool to address the need for a more convenient and efficient way to manage our data on the chain or other components. The purpose of venus-tool is to offer users of venus a comprehensive and user-friendly management interface for handling settings and data related to chain services, deal services, and power services. It also aims to resolve any conflicts arising from the separation of chain services and users.

This tool is currently under development. We welcome everyone to suggest new features, raise issues, and submit pull requests. If there are any features that you would like to see, please let us know.

## Features

- Overview of your assets and areas of interest
- Access and manage your messages
- View and interact with your sealing threads
- Review your mining records
- Monitor and handle your deals



## Usage

### Install And Run

#### Build from source

Just git clone the repo and make
```shell
make
```

#### Launch Up

You can run the binary directly
```shell
./venus-tool run \
--node-api=/ip4/{NODE_IP}/tcp/3453 \
--msg-api=/ip4/{MESSAGER_IP}/tcp/39812 \
--market-api=/ip4/{MARKET_IP}/tcp/41235 \
--miner-api=/ip4/{MINER_IP}/tcp/12308 \
--wallet-api={WALLET_TOKEN}:/ip4/{WALLET_IP}/tcp/5678/http \
--auth-api=http://{AUTH_IP}:8989 \
--damocles-api=/ip4/{DAMOCLES_MANAGER_IP}/tcp/1789 \
--common-token={CHAIN_SERVICE_TOKEN} \


```
tips: You can get `WALLET_API` from venus-wallet with `venus-wallet auth api-info` 

##### With Docker

build a docker image or pull "filvenus/venus-tool" from docker hub. 
```shell
make docker
```

run docker container
```shell
docker run -d filvenus/venus-tool:latest \
run \
--node-api=/ip4/{NODE_IP}/tcp/3453 \
--msg-api=/ip4/{MESSAGER_IP}/tcp/39812 \
--market-api=/ip4/{MARKET_IP}/tcp/41235 \
--miner-api=/ip4/{MINER_IP}/tcp/12308 \
--wallet-api={WALLET_TOKEN}:/ip4/{WALLET_IP}/tcp/5678/http \
--auth-api=http://{AUTH_IP}:8989 \
--damocles-api=/ip4/{DAMOCLES_MANAGER_IP}/tcp/1789 \
--common-token={CHAIN_SERVICE_TOKEN} \
```

#### Dashboard

Access your dashboard by visit `http://localhost:8090 if you run it in your local machine.


### More
For more detail , run `venus-tool -h`.


## Development Plan


### Command Line



- [x] Message management
	- [x] query message
	- [x] send message
	- [x] replace message
- [x] Miner management
	- [x] create miner
	- [x] set ask
	- [x] query deadline
	- [x] change owner , worker, controller
	- [x] withdraw fund
	- [x] withdraw fund from market 
- [x] Multi Sign  management
	- [x] manage multi sign address
		- [x] query
		- [x] create
	- [x] proposal management
		- [x] propose
		- [x] cancel
		- [x] approve
	- [x] signer management
		- [x] add
		- [x] remove
		- [x] swap
	- [x] wallet management
		- [x] sign record query
		- [ ] sign offline
		- [ ] set sign filter 


### Web UI


- [ ] Summary
	- [x] All  available balance 
	- [x] Total Collateral
	- [x] Total Raw Byte Power
	- [x] Total Quality Adjust Power
	- [x] Gas Used 
	- [x]  Collateral In Miner
	- [x] Collateral In Market
	- [x] Mined Block Count Expect
	- [ ] Mined Block Count Actually
- [x] Asset
	- [x] Miners
      - [ ] Set your miner about ask, worker, controller, beneficiary and so on
	- [x] Wallet Addresses
      - [x] Gas setting used when sending messages
- [ ] Message
	- [x] Message List
		- [x] Filt and sort msg by state
	- [ ]  Send
		- [x] Send message
		- [ ] Method Selector
	- [ ] Replace message
	- [x] Mark Bad message
		- [x] Basic
		- [x] Markbad batch
- [ ] Sealing
	- [x] Thread list
	- [x]  Stop thread
	- [x] Start thread
	- [x] Abort thread
	- [ ] Set state of thread
	- [ ] Time used
- [ ] Mine
	- [x] Mined record
- [ ] Deal
	- [x] Deal List
	- [ ] Publish  deal
	- [ ] Publish deal batch
- [x] Search
	- [x] Message detail
	- [x] Miner Detail
	- [x] Deal Detail
	- [x] Wallet address detail


## Preview

No online demo available yet. You can run it locally to have a try.
![Preview](https://github.com/filecoin-project/venus/assets/55120714/98b812be-526d-4a46-93e8-8ae193c58540)
