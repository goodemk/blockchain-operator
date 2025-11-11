const Web3 = require('web3');
const contract = require('truffle-contract');
const assert = require('assert');
const fs = require('fs');
var childProcess = require('child_process');

class Racecourse {

    init(url, user, password, eventListener, address, statusFunction) {
        let web3 = new Web3(new Web3.providers.HttpProvider(url, 0, user, password));
        return new Promise((resolve, reject)=> {
            if(web3.isConnected()) {
                statusFunction('Connection successful');
                this.accounts = web3.eth.accounts
                let raceContract = contract({
                    abi: JSON.parse(fs.readFileSync('../contracts/Race.abi', 'utf8')),
                    unlinked_binary: fs.readFileSync('../contracts/Race.bin', 'utf8')
                });
                raceContract.setProvider(web3.currentProvider);

                raceContract.synchronization_timeout = 60000; // 60 seconds
                raceContract.defaults({
                    gas: 5000000,
                    gasPrice: 0
                });

                if(address) {
                    statusFunction('Contract already deployed');
                    this.raceContract = raceContract.at(address).then((instance) => {
                        statusFunction('Using contract at address ' + address);
                        this.contractInstance = instance;
                        web3.eth.getBlockNumber((err, blockNumber) => {
                            this.listenForEvents(eventListener, blockNumber).then((events) => {
                                this.events = events;
                                resolve();
                            });
                        });
                    }).catch((error) => {
                        reject('Could not find contract at address ' + address);
                    });
                } else {
                    statusFunction('Deploying contract...');
                    this.raceContract = raceContract.new({from: web3.eth.accounts[0], gasPrice: 0, gas: 5000000}).then((instance) => {
                        statusFunction('Contract successfully deployed')
                        this.contractInstance = instance;
                        web3.eth.getBlockNumber((err, blockNumber) => {
                            this.listenForEvents(eventListener, blockNumber).then((events) => {
                                this.events = events;
                                resolve();
                            });
                        });
                    }).catch((err) => {
                        reject('Error deploying contract: ' + err);
                    });
                    
                    // Send a transaction to force the block to be mined with the contract
                    setTimeout(() => {
                        web3.eth.sendTransaction({from: web3.eth.accounts[0], gasPrice: 0, gas: 5000000});
                    }, 1000);
                }
    
            } else {
                reject('Unable to connect to ' + url);
            }
        });
    };

    listenForEvents(eventListener, skipBlock) {
        return new Promise((resolve, reject) => {
            try {
                let betPlacedEvent = this.contractInstance.betPlaced({}, { fromBlock: 'latest', toBlock: 'latest' });
                betPlacedEvent.watch((err, result) => {
                    if(err) {
                        console.log('Event watch error (betPlaced):', err.message || err);
                        return;
                    }
                    if(result && result.blockNumber != skipBlock) {
                        eventListener('Bet placed');
                    }
                });

                let playersReadyToRaceEvent = this.contractInstance.playersReadyToRaceChanged({}, { fromBlock: 'latest', toBlock: 'latest' });
                playersReadyToRaceEvent.watch((err, result) => {
                    if(err) {
                        console.log('Event watch error (playersReadyToRaceChanged):', err.message || err);
                        return;
                    }
                    if(result && result.blockNumber != skipBlock) {
                        eventListener('Players ready to race changed');
                    }
                });

                let raceFinishedEvent = this.contractInstance.finishedRace({}, { fromBlock: 'latest', toBlock: 'latest' });
                raceFinishedEvent.watch((err, result) => {
                    if(err) {
                        console.log('Event watch error (finishedRace):', err.message || err);
                        return;
                    }
                    if(result && result.blockNumber != skipBlock) {
                        eventListener('Finished race');
                    }
                });

                resolve([betPlacedEvent, playersReadyToRaceEvent, raceFinishedEvent]);
            } catch(error) {
                console.log('Warning: Could not setup event watching:', error.message || error);
                resolve([]);
            }
        });
    };

    stopWatching() {
        for(let event of this.events) {
            event.stopWatching();
        }
    }

    getState() {
        let raceContractInstance = this.contractInstance;
        let contractState = {};
        return raceContractInstance.horseCount().then((horseCount) => {
            let horsePromises = [];
            for (var i = 0; i < horseCount; i++) {
                horsePromises.push(raceContractInstance.horses(i));
            }
            return Promise.all(horsePromises).then((horses) => {
                contractState.horses = [];
                for (var i = 0; i < horseCount; i++) {
                    contractState.horses.push({
                        index: horses[i][0].toString(),
                        name: horses[i][1]
                    });
                }
                return raceContractInstance.betCount();
            });
        }).then((betCount) => {
            let betPromises = [];
            for (var i = 0; i < betCount; i++) {
                betPromises.push(raceContractInstance.bets(i));
            }
            return Promise.all(betPromises).then((bets) => {
                contractState.bets = [];
                for (var i = 0; i < betCount; i++) {
                    contractState.bets.push({
                        index: bets[i][0].toString(),
                        player: bets[i][1],
                        amount: bets[i][2].toString(),
                        playerReadyToRace: bets[i][3]
                    });
                }
                return raceContractInstance.playersReadyToRace();
            });
        }).then((playersReadyToRace) => {
            contractState.playersReadyToRace = playersReadyToRace.toString();
            return raceContractInstance.raceFinished();
        }).then((raceFinished) => {
            contractState.raceFinished = raceFinished;
            return raceContractInstance.jackpot();
        }).then((jackpot) => {
            contractState.jackpot = jackpot.toString();
            return raceContractInstance.winnerHorse();
        }).then((winnerHorse) => {
            contractState.winnerHorse = winnerHorse.toString();
            return contractState;
        }).catch((error) => {
            console.log(error);
        });
    };

    placeBet(index, amount, account) {
        return this.contractInstance.placeBet(index, amount, {from: account, gas: 5000000});
    }
    
    playerReadyToRace(account) {
        return this.contractInstance.playerReadyToRace({from: account, gas: 5000000});
    }

}

module.exports = Racecourse;