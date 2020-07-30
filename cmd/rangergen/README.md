## Ranger: generate code suitable for embedding into protobufs

Ranger generates marshaled structs in raw bytes in a way that lets them be
encapsulated on top of each other in a structural fashion, which makes them
ideal for encapulating inside protobuf transmissions. Each method generated
works with each other across types as well as with others that implement the
`ranger.Marshaler` interface, which is eerily similar to the gogoproto
marshaling interface, without depending on a second code generator.

Ranger's generation rules are fairly simple and largely encompassed around
structure; integer types supported are the `uint`s of various sizes which we
pass to `binary.Uvarint` for generation of both the integers themselves as
data, as well as array lengths for array structure data.

## Building/Running

It's simplest to run `rangergen` as a part of `go run`, but you can optionally compile it as well:

- `go run ./cmd/rangergen --help`
or
- `go install ./cmd/rangergen/... && rangergen --help`

## Using

`rangergen` will be used here -- see `Building/Running` for other methods of invocation.

`rangergen` generates test and fuzz code along with its generated corpus. To
suppress this; see the options.

`rangergen` supports making parent directories if necessary.

`rangergen definition.yaml path/prefix/to/file`

Will generate:

* `path/prefix/to/file.go` -- generated corpus
* `path/prefix/to/file_test.go` -- test files that test randomized i/o through the system briefly
* `path/prefix/to/file_fuzz.go` -- supporting tooling for
  [go-fuzz](https://github.com/dvyukov/go-fuzz). Read go-fuzz's linked docs for
  more information.

## Format

Fields are specified in a YAML format that you can see here:

```yaml
# package name
package: ledger
#
# definition of types. each definition consists of a Type name as map key and
# properties as values. Some properties, like `fields`, are required but others
# can be specified as well. I had a host of validations here that I thought
# could be implemented at the struct level for length, even post-marshal hooks
# like encryption.
#
# for now, the `fields` key is more or less what is defined.
##
types:
  Transactions:
    ##
    # comment is a comment that goes with the type, in golint-compatible format.
    ##
    comment: "is for holding a list of transaction structs."
    fields:
      - name: Transactions
        ##
        # structure_type is the type of structure -- this goes into how we
        # will marshal it, with prefixed length in the array case, or
        # pre-calculated size based on type size in the scalar one. The
        # options are scalar and array -- but I do have support for a map
        # type in the spec that relies on another field to set the map key
        # type. We do not need this right now according to
        # ledger/transaction.go.
        #
        # The default structure_type is `scalar`.
        ##
        structure_type: array
        ##
        # The value_type is the type of the actual item. Together with
        # structure_type, we can safely marshal structures with special
        # needs.  If a type is not natively supported it must conform to a
        # json.Marshaler style interface that we need to define out-of-band
        # to support codegen.
        ##
        value_type: Transaction
        require:
          max_length: 20
        ##
        # comment, like the other attributes is golint-compatible and belongs
        # to the struct member's documentation here.
        ##
        comment: "is the list of transactions"
  Transaction:
    fields:
      - name: Version
        value_type: uint8
        require:
          static: true
      - name: Body
        value_type: TransactionBody
        ##
        # interface means "this conforms to an interface which could be N
        # types"... the input is read from the head and used to unmarshal the
        # rest. the individual types are expected to have implementations for
        # the marshal functions we implement -- e.g., have been at least
        # partially generated or emulated generation for intricate needs.
        ##
        interface:
          # This is TxType() as a part of the TransactionBody interface; not
          # the type name. this could be pre-defined as a bit that conforms
          # to a codegen interface, instead of calling it TxType we coudl
          # call TransactionBodyType() or something. Then this yaml field
          # could focus on concrete types (and thus size).
          output: TxType
          # I know this is TxType under the hood, but it resolves to uint8 --
          # I'm not sure we want to be in the type management business this deep.
          input: uint8
          cases:
            - TxTypeTransfer: TransferTransaction
            - TxTypeGenesis: GenesisTransaction
            - TxTypeGlobalConfig: GlobalConfigTransaction
            - TxTypeEscrowOpen: EscrowOpenTransaction
        require:
          max_length: 20
      - name: Flags
        value_type: uint16
        require:
          static: true
  TransferTransaction:
    fields:
      - name: Inputs
        structure_type: array
        value_type: TransactionInput
        ##
        # matching rules for validations, other things that could be here:
        #
        # - enums
        ##
        match:
          ##
          # this matches a length of a field inside the struct. not sure if
          # we should require the pair definition but it's there for
          # posterity for now.
          ##
          length_of_field: Witnesses
        require:
          max_length: 20
      - name: Outputs
        structure_type: array
        value_type: TransactionOutput
        require:
          max_length: 20
      - name: Witnesses
        structure_type: array
        value_type: TransactionWitness
        match:
          length_of_field: Inputs
        require:
          max_length: 20
      - name: LockTime
        value_type: uint32
        marshal: false # this field is not marshaled
  EscrowOpenTransaction:
    fields:
  TransactionInput:
    fields:
      - name: Outpoint
        value_type: Outpoint
        require:
          max_length: 20
        ##
        # embedded embedds a struct into another as
        # Outpoint is in TransactionInput in
        # ledger/transaction.go.
        ##
        embedded: true
      - name: ScriptSig
        value_type: "[]byte"
        require:
          max_length: 520
      - name: SequenceNo
        value_type: uint32
  ##
  # this is my attempt to model Outpoint
  ##
  Outpoint:
    fields:
      - name: PreviousTx
        value_type: "[]byte"
        ##
        # this could be for validation requirements.
        # other things that could be here:
        # - format of data (e.g., gzip, or some packet format, etc. basically a mime type)
        ##
        require:
          length: 32
      - name: Index
        value_type: uint8
  TransactionOutput:
    fields:
      - name: Value
        value_type: uint32
      - name: ScriptPubKey
        value_type: "[]byte"
        require:
          max_length: 20
  TransactionWitness:
    fields:
      - name: Data
        structure_type: array
        value_type: "[]byte"
        require:
          max_length: 20
        item_require:
          max_length: 20
  GenesisTransaction:
    fields:
      - name: Outputs
        structure_type: array
        value_type: TransactionOutput
        require:
          max_length: 20
  GlobalConfigTransaction:
    fields:
      - name: ActivationBlockHeight
        value_type: uint64
      - name: ScalarUpdates
        structure_type: array
        value_type: GlobalConfigScalarUpdate
        require:
          max_length: 20
      - name: ListUpdates
        structure_type: array
        value_type: GlobalConfigListUpdate
        require:
          max_length: 20
      - name: SigPublicKey
        value_type: "[]byte"
        require:
          max_length: 20
      - name: Signature
        value_type: "[]byte"
        require:
          max_length: 20
  GlobalConfigScalarUpdate:
    fields:
      - name: Key
        value_type: string
        require:
          max_length: 20
      - name: Value
        value_type: "[]byte"
        require:
          max_length: 20
  GlobalConfigListUpdate:
    fields:
      - name: Key
        value_type: string
        require:
          max_length: 20
      - name: Deletions
        structure_type: array
        value_type: uint64
        require:
          max_length: 20
      - name: Insertions
        structure_type: array
        value_type: GlobalConfigListInsertion
        require:
          max_length: 20
  GlobalConfigListInsertion:
    fields:
      - name: Index
        value_type: uint64
      - name: Value
        value_type: "[]byte"
        require:
          max_length: 20

```

You can see additional examples in `/ranger/testdata` off the root of the tree.
