# Cache selection

When dealing with a request the publisher has to select a set of caches to serve
the chunks for the given byte ranges from one of the escrows that could cover
the content.

This selection should be designed to place the same chunks on the same caches
(expanding the number in rotation as demand for the chunks grows), and to reuse
the same caches for consecutive chunk-groups served to a single client (so that
connection reuse and pipelining can improve performance, once implemented).

## Design proposal: treat this as a multi-level consistent hashing problem

### Design goals medium term

A) globally we want all ranges available in each region.
B) a single bundle needs to be split over multiple caches for puzzle security (is this true?)
C) if a cache is failed or degraded we want to keep most other caches still serving the same chunks
D) we don't want any one client to be asked to use all the caches - there is a connection overhead per cache
   as well as TCP inefficiencies when too little data is sent on a pipe, so enough caches to spread load and
   meet security requirements, but no more.
E) if there are many more caches in that region, we want other clients to get spread to other caches so that
   we are not unfairly paying just a small fraction of escrow participants.
F) if resources are not available close to the client we should fall back globally

### Specific proposal v0

See CARP or the [maglev hashing system](http://static.googleusercontent.com/media/research.google.com/zh-TW//pubs/archive/44824.pdf)
for the inspiration here.

* Perform maglev permutation generation for the caches
  * these can be shared/cached across different escrows
* Precalculate a global lookup table mapping to a single cache in the escrows
  list of caches
* In bundle generation we hash the (path + content byte/ chunk offset) to get a lookup
  key into the lookup table From that cache we take the next bundle-chunk-count
  caches to satify the request.
  * path is used to ensure that even a publisher with only very short files
    will utilise all the caches in the escrow.
  * bundle offset is used to ensure a publisher with a hot long file won't end
    up over-using just a small subset of the caches in the escrow
  * other corner cases are not addressed (hot short/regular length files and
    capping client cache connections to small numbers)
  * if some of the lookup result is unusable (e.g. the client has signalled that
    it cannot use one or more cache, or the publisher has decided one ore more
    cache is unusuable, or some cache isn't responding to the publisher etc),
    then we just continue further in the list of caches to pick another cache.
  * If less than chunks-per-bundle caches can be used, an error results.
  * This v0 solution has the property that clients will talk to all caches in an
    escrow, so large escrows or low data rates will prevent TCP opening wide
    windows; see v1.

### Specific proposal v1 - add in managing subsets of cache per client (goal D)

Rather than including every cache in the lookup table, we generate many lookup
tables each with some optimal number of caches that we would like the clients to
be maintaining connections to under ideal conditions. This is done by taking the
first N such caches, generating one lookup table, and then shuffling caches out
and in in order. Consider a 6 cache scenario: A B C D E F.

An initial table might be

0 C
1 D
2 A
3 B
4 .
5 .
...

The rotations can then be arranged like so - for single-shuffle, complete rotations.

\ i j k l m n | o p q r s t * (start)
0 D D D D B B | B B F F F F D
1 C C C A A A | A E E E E C C
2 A E E E E C | C C C A A A A
3 B B F F F F | D D D D B B B
4 .

Note that we stop at the point where all the caches have rotated back in, as
that would be worst case misses with everything looking up in different places.

These then allow us to have multiple lookup tables that will collide on lookups
and thus share cache placements (without stateful cache chunk placement which
longer we may well want eventually anyway - but that comes with consequences
on performance and uptime from tracking).

For dealing with failed caches, the v0 algorithm gets used - this will pick
caches from the next rotation over, and if rotation sharing is being used these
will be cache hits.

To maximise hit rate we select a rotation based on hash of the object path.

We can trade off that with greater use of rotations to spread objects over more
rotations; this could be dynamic based on load metrics or configuration on
routes or - lots of options. ... or we go straight to a placement management
system and track load and performance and cache utilisation as a direct online
scheduling problem and balancing problem.

### Specific proposal v2 - add in region awareness

* Create an additional layer of regions over the other lookup tables
* Each regional lookup table is built as the previous global system was
* If there are insufficient caches in a region, the lookup table is built just using the nearest caches
* If a clients region cannot be determined a hash(client pubkey + client net
  address) can be used to key into the set of regional lookup tables to get an
  arbitrary region
