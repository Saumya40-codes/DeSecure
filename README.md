## Decentralized Rights Management CLI
A command-line tool for managing digital content rights on a decentralized blockchain network with IPFS integration. This creates a prototypical blockchain to play with


### Overview
This project implements a decentralized digital rights management (DRM) system that allows users to:

- Upload digital content to IPFS
- Register content ownership on a blockchain
- Purchase content licenses from other users
- Validate and verify content access rights
- View detailed blockchain and transaction history

The system uses a proof-of-authority consensus mechanism with multiple validators to ensure transaction integrity and prevent unauthorized access to digital assets.

### Features

- Asset Management: Upload, list, and access digital assets
- Blockchain Integration: All transactions recorded on a transparent blockchain
- Decentralized Storage: Content stored on IPFS for censorship resistance
- License Verification: Cryptographic validation of access rights
- P2P Network: Decentralized node discovery and communication
- Configurable Licenses: Support for different license types

### Technical Components

- Golang-based CLI application
- Peer-to-peer networking using libp2p
- Content addressed storage with IPFS
- BadgerDB for local blockchain and key storage
- ECDSA for transaction signing and verification
- Consensus protocol for transaction validation

### Setup Requirements

- Go: https://go.dev/doc/install
- Docker (to run IPFS node) (it should be running with docker or its direct installation)
