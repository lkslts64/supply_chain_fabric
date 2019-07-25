/*
SPDX-License-Identifier: Apache-2.0
*/

/*
 * This application has 6 basic steps:
 * 1. Select an identity from a wallet
 * 2. Connect to network gateway
 * 3. Access PaperNet network
 * 4. Construct request to issue commercial paper
 * 5. Submit transaction
 * 6. Process response
 */

'use strict';

// Bring key classes into scope, most importantly Fabric SDK network class
const fs = require('fs');
const yaml = require('js-yaml');
const { FileSystemWallet, Gateway } = require('fabric-network');
const Client = require('fabric-client')
//const CommercialPaper = require('../contract/lib/paper.js');

// A wallet stores a collection of identities for use
//const wallet = new FileSystemWallet('../user/isabella/wallet');
const wallet = new FileSystemWallet('../identity/user/loukas/wallet');

// Main program function
async function main() {
  // A gateway defines the peers used to access Fabric networks
  const gateway = new Gateway();

  // Main try/catch block
  try {

    // Specify userName for network access
    // const userName = 'isabella.issuer@magnetocorp.com';
    const userName = 'Admin@org1.example.com';

    // Load connection profile; will be used to locate a gateway
    let connectionProfile = yaml.safeLoad(fs.readFileSync('../gateway/networkConnection.yaml', 'utf8'));
    //let client = Client.loadFromConfig('../gateway/networkConnection.yaml')

    // Set connection options; identity and wallet
    let connectionOptions = {
      identity: userName,
      wallet: wallet,
      discovery: { enabled:false, asLocalhost: true }
    };

    // Connect to gateway using application specified parameters
    console.log('Connect to Fabric gateway.');

    await gateway.connect(connectionProfile, connectionOptions);

    // Access PaperNet network
    console.log('Use network channel: mychannel.');

    const network = await gateway.getNetwork('mychannel');

    // Get addressability to commercial paper contract
    console.log('Use org.papernet.commercialpaper smart contract.');

    const contract = await network.getContract('scthreediff6');

    // issue commercial paper
    console.log('Submit commercial paper issue transaction.');

    //const issueResponse = await contract.submitTransaction('deliverCrude','Crude1','207','200','orgDriller','2002-10-02T10:00:00-05:00','orgDriller','org1','342352','2003-10-02T10:00:00-05:00');
	const issueResponse = deliverCrudeRand(contract,1);
    console.log('Process issue transaction response.');

    //let paper = CommercialPaper.fromBuffer(issueResponse);

    //console.log(`${paper.issuer} commercial paper : ${paper.paperNumber} successfully issued for value ${paper.faceValue}`);
    console.log('Transaction complete.');
    console.log(issueResponse);

  } catch (error) {

    console.log(`Error processing transaction. ${error}`);
    console.log(error.stack);

  } finally {

    // Disconnect from the gateway
    console.log('Disconnect from Fabric gateway.')
    gateway.disconnect();

  }
}


function deliverCrude(contract,crude_num,value,quant,owner,estTime,startLoc,dest,vessel_id) {
	return contract.submitTransaction('deliverCrude','Crude'+crude_num,value,quant,'org'+owner,estTime,dest,vessel_id)
}


function deliverCrudeRand(contract,crude_num) {
	let value = Math.floor(Math.random()*101) +1;
	let quant = Math.floor(Math.random()*101) +1;
	let owner = 'org1';

	let dur = Math.floor(Math.random()*101) +1;
	let time = new Date();
	time.setSeconds(time.getSeconds() + dur)
	let estTime = time.toISOString();
	let startLoc = owner;
	let dest = 'org3';
	let vessel_id = Math.floor(Math.random()*1001) +1;
	return contract.submitTransaction('deliverCrude','Crude'+crude_num,value,quant,owner,estTime,startLoc,dest,vessel_id)
}

function refineRand(contract,fuel_num,crude_num) {
	let value = Math.floor(Math.random()*101) +1;
	let quant = Math.floor(Math.random()*101) +1;
	let owner = 'org3';
	let density = Math.floor(Math.random()*101) +1;
	let type = 'fuel';
	return contract.submitTransaction('refine','Fuel'+fuel_num,value,quant,owner,density,type,'Crude'+crude_num,(new Date).toISOString())
}

function addFuelOrderRand(contract,fuelOrder_num,fuel_num) {
	let value = Math.floor(Math.random()*101) +1;
	let quant = Math.floor(Math.random()*101) +1;
	let owner = 'org3';
	let rcoin = Math.floor(Math.random()*2);
	let dest;
	if (rcoin == 1) 
		dest = 'org5';
	else if (rcoin == 2) 
		dest = 'org6';
	return contract.submitTransaction('addFuelOrder','FuelOrder'+fuelOrder_num,value,quant,owner,dest,'Fuel'+fuel_num,(new Date()).toISOString())
}

function deliverFuelRand(contract,plan_num,fuelOrders) {
	let trackid = Math.floor(Math.random()*10001) +1;
	let i;
	for (i = 0; i < fuelOrders.length; i++) {
	}
}

main().then(() => {

  console.log('Issue program complete.');

}).catch((e) => {

  console.log('Issue program exception.');
  console.log(e);
  console.log(e.stack);
  process.exit(-1);

});
