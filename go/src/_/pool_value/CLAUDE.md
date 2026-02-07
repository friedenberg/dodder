# pool_value

Generic sync.Pool wrappers with automatic reset functionality.

## Types

- `poolValue[T]`: Generic pool with custom reset logic
- `poolSlice[T, []T]`: Specialized pool for slices with automatic truncation

## Functions

- `Make[T](construct, reset)`: Create pool with constructor and reset functions
- `MakeSlice[T, []T]()`: Create slice pool with automatic [:0] reset

## Methods

- `Get()`: Get item from pool (constructs if needed)
- `Put(item)`: Return item to pool after resetting

Provides type-safe pooling with automatic cleanup, preventing memory leaks
and reducing allocations for frequently-used objects.
