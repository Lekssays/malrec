#!/usr/bin/python3
import argparse
import json
import subprocess
import os

from string import Template


CORE_PEER_ID = os.getenv('CORE_PEER_ID')
CORE_PEER_ADDRESS = os.getenv('CORE_PEER_ADDRESS')
CORE_PEER_TLS_ROOTCERT_FILE = os.getenv('CORE_PEER_TLS_ROOTCERT_FILE')
ORDERER_CA = os.getenv('ORDERER_CA')
BACKUP_FOLDER = '/backup'
BASE_COMMAND = 'peer chaincode query -o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile {} -C mychannel -n backup --peerAddresses {} --tlsRootCertFiles {} -c '.format(ORDERER_CA, CORE_PEER_ADDRESS, CORE_PEER_TLS_ROOTCERT_FILE)

def parse_args():
    parser = argparse.ArgumentParser(formatter_class=argparse.ArgumentDefaultsHelpFormatter)
    parser.add_argument('-o','--operation',
                        dest = "operation",
                        help = "1) To download all device backups\n\
                                2) To download valid backups in timestamp range (specify also start in -s argument and end timestamp in -e argument)\
                                3) To download a specific backup by backupID (specify backup ID in -b argument)",
                        default = "1",
                        required = True)
    parser.add_argument('-s','--start',
                        dest = "start",
                        help = "start timestamp\n",
                        default = "",
                        required = False)
    parser.add_argument('-e','--end',
                        dest = "end",
                        help = "end timestamp",
                        default = "",
                        required = False)
    parser.add_argument('-b','--backupID',
                        dest = "backupID",
                        help = "BackupID to download",
                        default = "",
                        required = False)                        
    return parser.parse_args()


def download_backup(hash: str, timestamp: str):
    command = 'ipfs get -o {}/{}"_downloaded.tar.gz" {}'.format(BACKUP_FOLDER, timestamp, hash)
    os.system(command)


def get_backup(backupID: str):
    t = Template('{"function":"QueryBackup","Args":["$backupID"]}')
    payload = t.substitute(backupID=backupID)
    command = BASE_COMMAND + "'{}'".format(payload)
    backup = subprocess.check_output(command, shell=True)
    backup = json.loads(backup.decode().strip())
    download_backup(hash=backup['hash'], timestamp=str(backup['timestamp']))

def get_backups():
    t = Template('{"function":"QueryBackupsByDeviceID","Args":["$deviceID"]}')
    payload = t.substitute(deviceID=CORE_PEER_ID)
    command = BASE_COMMAND + "'{}'".format(payload)
    backups = subprocess.check_output(command, shell=True)
    backups = json.loads(backups.decode().strip())
    for backup in backups:
        download_backup(hash=backup['hash'], timestamp=str(backup['timestamp']))


def get_backups_by_timestamps(start: str, end: str):
    t = Template('{"function":"QueryBackupsByTimestamps","Args":["$deviceID", "$start", "$end"]}')
    payload = t.substitute(deviceID=CORE_PEER_ID, start=start, end=end)
    command = BASE_COMMAND + "'{}'".format(payload)
    backups = subprocess.check_output(command, shell=True)
    backups = json.loads(backups.decode().strip())
    for backup in backups:
        download_backup(hash=backup['hash'], timestamp=str(backup['timestamp']))


def main():
    # EXAMPLES
    # python3 engine.py -o 1
    # python3 engine.py -o 2 -s 1621597001 -e 1621597001
    # python3 engine.py -o 3 -b peer0.org1.example.com_1621596959

    operation = parse_args().operation
    if operation == "1":
        get_backups()
    elif operation == "2":
        if parse_args().start and parse_args().end:
            get_backups_by_timestamps(start=parse_args().start, end=parse_args().end)
        else:
            print("Please specify start and end timestamps")
    elif operation == "3":
        if parse_args().backupID:
            get_backup(backupID=parse_args().backupID)
        else:
            print("Please specify backupID")
    else:
        print("Invalid operation")

if __name__ == '__main__':
    main()
