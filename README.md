<h1 align="center">Venus Tool</h1>

<p align="center">
 <a href="https://github.com/ipfs-force-community/venus-tool/actions"><img src="https://github.com/ipfs-force-community/venus-tool/actions/workflows/build_upload.yml/badge.svg"/></a>
 <a href="https://codecov.io/gh/ipfs-force-community/venus-tool"><img src="https://codecov.io/gh/ipfs-force-community/venus-tool/branch/master/graph/badge.svg?token=J5QWYWkgHT"/></a>
 <a href="https://goreportcard.com/report/github.com/ipfs-force-community/venus-tool"><img src="https://goreportcard.com/badge/github.com/ipfs-force-community/venus-tool"/></a>
 <a href="https://github.com/ipfs-force-community/venus-tool/tags"><img src="https://img.shields.io/github/v/tag/ipfs-force-community/venus-tool"/></a>
  <br>
</p>

## Why

We need a more convenient and efficient way to manage our data on chain or other component, so we create this tool.
venus-tool hopes to provide users of venus with a more convenient and complete management interface for managing the settings and data of users on chain services, deal services, and power services. At the same time, it reconciles the contradictions caused by the separation of chain services and users

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
--miner-api=/dns/miner/tcp/12308 \
--wallet-api={WALLET_TOKEN}:/ipv4/{WALLET_IP}/tcp/5678/http \
--auth-api=http://{AUTH_IP}:8989 \
--damocles-api=/ip4/{DAMOCLES_MANAGER_IP}/tcp/1789 \
--common-token={CHAIN_SERVICE_TOKEN} \





```
tips: You can get `WALLET_API` from venus-wallet with `venus-wallet auth api-info` 

#### With Docker

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
--miner-api=/dns/miner/tcp/12308 \
--wallet-api={WALLET_TOKEN}:/ipv4/{WALLET_IP}/tcp/5678/http \
--auth-api=http://{AUTH_IP}:8989 \
--damocles-api=/ip4/{DAMOCLES_MANAGER_IP}/tcp/1789 \
--common-token={CHAIN_SERVICE_TOKEN} \
```

#### Dashboard

venus-tool provides a dashboard , you can access it by `http://localhost:8090 if you run it in docker.


### More
For more detail , run `venus-tool -h`.
