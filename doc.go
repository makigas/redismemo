/*
Package redismemo provides a basic memoization system around Redis.

It works by wrapping a go-redis/v9 client and providing a simple function
that can be used to fetch a value from Redis given its key, as well as the
compute function to be issued if the value is not found in the Redis
instance. If the value has to be computed, it is also set in the Redis
instance using the given TTL, which comes handy for caches.
*/
package redismemo
