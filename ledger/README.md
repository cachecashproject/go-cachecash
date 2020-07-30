# ledger

## Notes

- We support only native P2WPKH transactions (as described in BIP-0142), not P2PKH transactions, P2WPKH-in-P2SH, or
  other variants.  P2SH transactions are not currently supported but will be in a future iteration, in order to support
  cross-chain atomic swaps.

## P2WPKH transactions

A P2WPKH output script looks like

    OP_0 [20-byte public key hash]

    0x00 0x14 [20 bytes]

The constant 0 is the witness program version.  It is followed by a push of the public key hash.  The output address
includes an address version, which identifies the address as a P2WPKH address on a particular network (testnet, goodnet,
or mainnet).

An input spending a P2WPKH output must have a two witness values attached (`[signature] [public key]`) and an
empty input script.  The empty input sript is replaced with

    OP_DUP OP_HASH160 [20-byte public key hash] OP_EQUALVERIFY OP_CHECKSIG

    0x76 0xa9 0x14 [20 bytes] 0x88 0xac

After the output script is evaluated, the stack will look like

    [20-byte public key hash]
    0

The address version indicates that witnesses are in use, so the 0 is consumed (indicating the witness program version);
because this is version 0, the VM checks

- that the input script is empty;
- that there is a single value left on the stack; and
- that there are two witness values

and then

- replaces the empty input script with the template above, consuming the stack; and
- pushes the two witness values onto the stack.

After this, script execution proceeds as normal.  If the public key in the witness data matches the hash embedded in the
address and the signature in the witness data is valid, execution succeeds and the output may be spent.
