# Errors

Golang errors with stacktraces.

Basically a trimmed down and cleaned up version of https://github.com/palantir/stacktrace.

## Why?

Because this could be hard to debug:

```
could not create service
```

Especially when compared to this:

```
could not create service
 --- at example_test.go:12 (NewController) ---
Caused by: could not create repository
 --- at example_test.go:18 (NewService) ---
Caused by: could not connect to database
 --- at example_test.go:23 (NewRepository) ---
```

## Usage

See the [tests](errors_test.go) for basic usage.

## License

The code is subject to the [MIT license](https://opensource.org/licenses/MIT).
