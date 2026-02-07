# pool

Generic sync.Pool wrappers with reset support.

## Key Types

- `pool[SWIMMER, SWIMMER_PTR]` - pointer-based pool wrapper
- `value[SWIMMER]` - value-based pool wrapper

## Key Functions

- `Make()` - creates pool with custom New/Reset functions
- `MakeWithResetable()` - creates pool for types implementing Resetable
- `MakeValue()` - creates value-based pool

## Key Methods

- `Get()` - retrieves element from pool
- `GetWithRepool()` - gets element with automatic return function
- `Put()` - returns element to pool (with reset)

## Features

- Type-safe generic pools
- Automatic reset on Put()
- Bespoke pools (bespoke.go) and fake pools (fake_pool.go) for testing
- FakePool with error injection via WithError wrapper
