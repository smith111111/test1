Bolt db implementation
========================

The `Datastore` interface allows for a pluggable wallet database. GalaxyOTC, for example, uses its own `Datastore` implementation
so that wallet data can be stored in the same database alongside the rest of GalaxyOTC data so that users need only make one backup.

Writing your own implementation is probably the best approach, however, this package does contain a workable `Datastore` implementation
using BoltDB. This is the datastore used by `NewDefaultConfig`.