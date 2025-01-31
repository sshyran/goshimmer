
# Layers
> DISCLAIMER: The described layers describe how things are implemented in GoShimmer. They might not reflect final Coordicide specification or implementations.



GoShimmer abstracts node functionalities into different layers. Similar to other architectures, upper layers build on the provided functionality of the layers below them. A layer is merely the concept of creating a clear separation of concerns.

Layers operate on payloads and it is up to the layer to react to the wanted payload types.

## Communication Layer
The communication layer is the most primitive layer, as its job is to simply form a graph made out of messages which contain payloads. As the name implies, messages are communicated/gossiped throughout the entire network. Think of it as the "physical layer" in the OSI-model. This layer forms a DAG made up from messages as a each message references two previous messages.

### Message
A message is a core data type which reflects a vertex in the communication layer [DAG](https://en.wikipedia.org/wiki/Directed_acyclic_graph).

It contains following properties:
* References to other messages
* Issuer's public key
* The issuance time of the message
* The message sequence number from the node which issued the message
* Payload which might be interpreted by upper layers
* The nonce which lets the message fulfill the PoW requirement
* A Signature signing all of the above fields.

A message is gossiped only when it becomes solid, meaning that its past history is known to the node. Messages currently must become solid within a 30 seconds time period, otherwise they are discarded.

Messages must also fulfill a PoW requirement currently which involves finding a nonce so that the hash of the fields of the message (minus the signature) has a certain amount of leading zeros. The PoW currently operates on `BLAKE2b_512`. Later the PoW requirement will be substituted by the actual Coordicide rate control mechanisms.

A message's byte layout is defined as:
```
parent_0<64bytes>
parent_1<64bytes>
issuer_public_key<32bytes>
issuance_time<int64,8bytes>
sequence_number<uint64,8bytes>
payload<variable-size, max 64KBs>
nonce<uint64,8bytes>
signature<64bytes>
```

### Payloads
As described above, a message contains a payload. In GoShimmer, there are 3 defined payload types, however, [more such types can be defined by developers seeking to implement their own application on top of the communication layer](https://github.com/iotaledger/goshimmer/wiki/How-to-create-a-simple-dApp).

| Type ID | Name | Purpose |
| -------- | -------- | -------- |
| 0     | Data     | Holds raw bytes without any further meaning     |
| 1     | Value Object     | Represents a value object on the value layer      |
| 111     | DRNG Object     | Represents a DRNG object for the DRNG layer     |

A payload's byte layout is defined as:
```
type<uint32-4bytes>
length<uint32-4bytes>
data<length bytes>
```

It is a upper layer's concern to listen for messages which hold specific payload types up on which the layer operates on.

### Tip-Selection
Since on the communication layer the payloads do not impose any restriction on the validity of a message, the tip-selection can simply operate on a pool of recent solid tips. However, in the future the tip-selector might impose certain restrictions such as below-max-depth checks and so on.

## Value Layer
The value layer operates solely on payloads of type value object. This layer has multiple responsibilities:
* Forming the ledger state
* Processing, validating and issuing transactions
* Conflict detection
* Conflict resolution via FPC
* Forming a DAG made up from value objects
* Tip-selection (on value object tips)

### Value Object
A value object is an object derived from a message containing a value object payload. It references two other value objects and contains one transaction.

The two references express vouching for the referenced value objects, meaning that they are seen as valid from the PoV of the value object which references them.

A value object is solid when its past cone is known and the contained transaction's inputs are known.

A transaction can occur in multiple value objects. In that case we speak of reattachments.

A value object's byte layout is defined as:
```
type<uint32-4bytes>
length<uint32-4bytes>
parent_0_reference<32bytes>
parent_1_reference<32bytes>
transaction<variable size>
```

### UTXO
GoShimmer uses unspent transaction outputs (UTXOs) as inputs for transactions. Meaning that inputs reference a specific UTXO ID which was generated by a previous transaction. The UTXO ID is made up from the destination address plus the hash of the transaction creating the output.

#### Inputs
As just described, an input is merely a reference to another transaction's output to a given address. GoShimmer currently works on addresses and unlike other cryptocurrencies does not yet have any concept of other unlocking mechanisms.

Addresses are BLAKE2b hashes of the corresponding Ed25519 and BLS public keys.

#### Outputs
An output encapsulates a destination address and a list of balances. A balance is an amount of tokens and a color.

##### Color & Coloring
The color of a balance is simply an array of 32 bytes and per default, tokens have a color of type "IOTA", where all 32 bytes are zero.

Creating a new color involves specifying the special color type "New" in a transaction output, where each byte in the color array is set to 255. Doing this will instruct GoShimmer to then color the specific balance on the specific address of the given output to the hash of the transaction creating the output. New colors therefore simply equal the transaction hash of the transaction creating the output (and are therefore unique).

Coloring tokens does neither increase or decrease the token supply. It is up to merchants and other systems to provide meaning to the colors of tokens.

Because of the way the coloring works, any holder of colored tokens can "shrink" the circulating amount of the specific colored token by simply creating outputs again with the special color type "New". Of course doing so will only effect the actual holder of the colored tokens but it is something to keep in mind when developing a system around colored tokens.

### Transaction
A transaction defines a transfer from UTXOs to new outputs.

It contains following properties:
* Inputs
* Outputs
* Payload
* Signatures

These properties minus the signatures make up the "essence" bytes of the transaction. Each signature signs the essence data. Therefore, if a transaction contains multiple inputs consuming UTXOs from an address, only one signature for that address needs to be provided.

A transaction is marked as solid when all of its referenced inputs are known.

A transactions’s byte layout is defined as:

```
inputs_count<uint32-4bytes>
 (per input)->
     address<33bytes>
     utxo_id<32bytes>
     
outputs_count<uint32-4bytes>
 (per output)->
     address<33bytes>
     balances_count<uint32-4bytes>
      (per balance)->
          value<int64-8bytes>
          color<32bytes>
          
payload_length<int32-8bytes>
payload<payload_length bytes>

(signatures (multiple occurrences in any order))->
    (ED25519)->
        signature_type(1 byte, value = 1)
        public_key<32bytes>
        signature<64bytes>
    (BLS)->
        signature_type(1byte, value = 2)
        public_key<128bytes>
        signature<64bytes>
signature_block_end<1byte, value 0>
```
 
The payload length inside the transaction can be max 65KBs.

### Parallel reality based ledger state
GoShimmer uses the parallel reality based ledger state [introduced by Hans in his blogpost series](https://medium.com/@hans_94488/a-new-consensus-the-tangle-multiverse-part-1-da4cb2a69772). Unlike the blogpost however, conflict resolution is done via [FPC](https://blog.iota.org/consensus-in-the-iota-tangle-fpc-b98e0f1e8fa). There is also no notion of mana as of yet or 'partial liking' other opinion realities.

Note that the reason the ledger state is implemented this way is to make it easier to extend it for the multiverse consensus in case it becomes a viable choice.

#### Branches / Realities
Realities are called branches in GoShimmer. A branch's ID is defined as a 32 byte long array. The master branch's ID is a 32 byte array where the first byte has the value 1.

Branches are created every time a conflict arises. The created branches have the IDs of the transactions which inflict/create the conflict.

Branches are in some sense a performance optimization as they allow to group transactions reflecting the same "reality" of ledger mutations as a group.

##### Conflict detection
As mentioned, transactions use UTXO as inputs. Every time a transaction consumes a given output, the consumer count on the output is incremented. If the consumer count goes above 1, a conflict arises, since multiple transactions are trying to consume/use the same UTXO (double spend).

Note that in the case that there is yet no finalized/confirmed consumer of the UTXO, the transactions forming the conflict set each spawn their own branch. Likewise, if a transaction is the first consumer of an output, it doesn't create a new branch (may however generate an aggregated branch).

##### Aggregated Branches
Since transactions are assigned to branches, transactions consuming outputs from transactions (which are non-conflicting) which reside in different branches, can create new so called aggregated branches. An aggregated branch is a virtual branch which reflects the change set of mutations of the ledger aggregated out of different branches.

#### Balances
Making up the balances of a given address involves querying the node for UTXOs which have as a target the given address.

### Consensus

#### FPC & FCoB
We define a conflict as consuming (i.e., spending) more than once an (unspent) output. 

Initially a node likes a transaction `v` that solidified at time `t` if in the interval `(−∞, t + avgNetworkDelay]` there are no spends from the same output (used as input).

If a conflicting transaction solidifies after that, we initially dislike it and add both (initially liked and initially disliked) conflicting transactions to the unresolved conflict set.

If no conflicting transactions solidify in the interval `(−∞, t + (2* avgNetworkDelay)]` we confirm (definitively like) `v`.

If `v_1,...,v_k` are all re-attachments of the same transaction, we either like all or none following the rule above.

Every `T` seconds a “Fast Probabilistic Consensus” (FPC) voting round is applied to every unresolved conflicting transaction in the Tangle:
 
+ each node queries a set of randomly chosen nodes about its unresolved conflicting transactions; 
+ the queried nodes send back their opinions on the requested transactions;

After the FPC is done the values `g(v)` of several transactions may change and then it stays the same forever. 
That is, after the vote a transaction is either confirmed (definitely liked) or rejected (definitely disliked) by a node, and this value will never change. Further discussion about monotonicity appears below.

#### Monotonicity
If `u` approves `v`, then `g(v) ≥ g(u)`, that is, if a node likes u then it likes any transaction `u` approves, and if the node dislikes `v` then it dislikes any transaction that approve `v`. 

Another way of saying it is that if we like `v` then we like all of its past cone, and if we dislike `v` we dislike all of its future cone.
The votes of any node will follow the monotonicity rule.
