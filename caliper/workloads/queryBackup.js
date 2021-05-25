'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class MyWorkload extends WorkloadModuleBase {
    constructor() {
        super();
    }
    
    async initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext) {
        await super.initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext);

        for (let i=0; i<this.roundArguments.backups; i++) {
            const backupID = `${this.workerIndex}_${i}`;
            console.log(`Worker ${this.workerIndex}: Creating backup ${backupID}`);
            const request = {
                contractId: this.roundArguments.contractId,
                contractFunction: 'CreateBackup',
                invokerIdentity: 'peer0.org1.example.com',
                contractArguments: [backupID,'QmdXYvmSEXrA9EoFBDQJRqrYiBLF6UB5o5M3pBSM4xJMuH'],
                readOnly: false
            };

            await this.sutAdapter.sendRequests(request);
        }
    }
    
    async submitTransaction() {
        const randomId = Math.floor(Math.random()*this.roundArguments.backups);
        const myArgs = {
            contractId: this.roundArguments.contractId,
            contractFunction: 'QueryBackup',
            invokerIdentity: 'peer0.org1.example.com',
            contractArguments: [`${this.workerIndex}_${randomId}`],
            readOnly: true
        };

        await this.sutAdapter.sendRequests(myArgs);
    }
    
    async cleanupWorkloadModule() {
        for (let i=0; i<this.roundArguments.backups; i++) {
            const backupID = `${this.workerIndex}_${i}`;
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