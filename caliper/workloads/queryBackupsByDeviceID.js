'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

const peers = ['peer0.org1.example.com', 'peer0.org2.example.com', 'peer0.org3.example.com'];

class MyWorkload extends WorkloadModuleBase {
    constructor() {
        super();
    }
    
    async initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext) {
        await super.initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext);

        for (let i=0; i<this.roundArguments.backups; i++) {
            const backupID = `BACKUP_${this.workerIndex}_${i}`;
            const peerId = Math.floor(Math.random() * (peers.length - 1));
            console.log(`Worker ${this.workerIndex}: Creating backup ${backupID} for peer ${peers[peerId]}`);
            const request = {
                contractId: this.roundArguments.contractId,
                contractFunction: 'CreateBackup',
                invokerIdentity: 'peer0.org1.example.com',
                contractArguments: [backupID, peers[peerId], 'QmdXYvmSEXrA9EoFBDQJRqrYiBLF6UB5o5M3pBSM4xJMuH'],
                readOnly: false
            };

            await this.sutAdapter.sendRequests(request);
        }
    }
    
    async submitTransaction() {
        const peerId = Math.floor(Math.random() * (peers.length - 1));
        const myArgs = {
            contractId: this.roundArguments.contractId,
            contractFunction: 'QueryBackupsByDeviceID',
            invokerIdentity: 'peer0.org1.example.com',
            contractArguments: [`${peers[peerId]}`],
            readOnly: true
        };

        await this.sutAdapter.sendRequests(myArgs);
    }
    
    async cleanupWorkloadModule() {
        for (let i=0; i<this.roundArguments.backups; i++) {
            const backupID = `BACKUP_${this.workerIndex}_${i}`;
            console.log(`Worker ${this.workerIndex}: Deleting backup ${backupID}`);
            const request = {
                contractId: this.roundArguments.contractId,
                contractFunction: 'DeleteBackup',
                invokerIdentity: 'peer0.org1.example.com',
                contractArguments: [backupID],
                readOnly: false
            };

            await this.sutAdapter.sendRequests(request);
        }
    }
}

function createWorkloadModule() {
    return new MyWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;