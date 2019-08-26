Utilzing blockchain and Hyperledger fabric to manage and monitor the supply chain.


Prerequisites:
 - Hyperledger fabric 1.4


In order to build the network (Debian platforms) :
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

1) clone this repo under fabric samples directory
2) copy chaincode directory under fabric-samples/chaincode/ 
3) navigate under first-network directory
4) ./byfn up
5) docker exec -it cli bash 
6) cd scripts && ./upgrade.sh 8.0 

In order to make transactions and query the network with the SDK:
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

1) navigate under app/application directory.
2) run: node addToWallet.js
3) node init.js
4) Now you are ready to transact with the blockchain. 
5) Run issue.js to update the blockchain and after serve.js to query the blockchain.



