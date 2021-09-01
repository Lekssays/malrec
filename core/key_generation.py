# https://pypi.org/project/bip-utils/
# https://pypi.org/project/eciespy/
import binascii
from bip_utils import Bip39MnemonicGenerator, Bip39WordsNum, Bip39SeedGenerator, Bip32

# Generate a mnemonic string of 15 words; the words are random.
mnemonic = Bip39MnemonicGenerator().FromWordsNumber(Bip39WordsNum.WORDS_NUM_15)
# print(mnemonic) 

# Seed generation; we must use BIP39. 
# We do not specify the language because it can be automatically detected
seed_bytes = Bip39SeedGenerator(mnemonic).Generate()
# print(seed_bytes)

# Get the BIP32 master key from the seed
bip32_master_key = Bip32.FromSeed(seed_bytes)
# print(bip32_master_key.PrivateKey().ToExtended())
# print(bip32_master_key.PublicKey().ToExtended())


# Get a key pair from the master key
# According to: 
# https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki#Specification_Wallet_structure
# Derivation path: tells how to derive a specific key withon a tree of keys
# https://river.com/learn/terms/d/derivation-path/
# m = master key
# m/0 = the first child of the master key
# We should not need hardened keys 
bip32_master_key = bip32_master_key.ChildKey(0)
#print(bip32_master_key.PrivateKey().ToExtended())
private_key = bip32_master_key.PrivateKey().Raw().ToHex()
# print(bip32_master_key.PrivateKey().Raw().ToHex())
public_key = bip32_master_key.PublicKey().RawUncompressed().ToHex()
#print(bip32_master_key.PublicKey().RawUncompressed().ToHex())

print(private_key)
print(public_key)
# print(private_key, file=open('key/private_key', 'w'))
# print(public_key, file=open('key/public_key', 'w'))
