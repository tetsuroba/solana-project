# Solana Project

## Description
This project mainly focuses on the Solana blockchain.
It supports the following features:
- [x] Create a wallet
- [x] Monitor a wallet and send notifications when a transaction is made
- [ ] Send a token swap transaction
- [ ] Scan for profitable wallets

## Prerequisites
- Go 1.21
- MongoDB 
- Docker

## Setup
1. Setup .env file according to .env.example
2. run `docker build -t my-go-app .`
3. run `docker run -p 8080:8080 my-go-app`
