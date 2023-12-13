# Fiber Functionality Exploration

## Requirements

- [ ] Default Values ❌
- [x] Form Binding Validation
- [x] Custom Form Binding Validation
- [x] URI (Path Param) Binding
- [x] Middlewares
- [x] Middleware execution chain
- [x] Recovery Middleware
- [x] Rewrite paths
- [x] Customizable Request JSON (❌) logger Middleware
- [x] Pprof
- [x] Route Groups
- [x] Expose Prometheus Metrics
- [x] Context Management -> Add Keys and Fork context

## Issues + Resolutions

1. Implementation for default values is missing, and the MR [#2699](https://github.com/gofiber/fiber/pull/2699) for the same seems to be stalled as well as sub par. This may require a custom implementation for now.
    - Added custom implementation for filling in defaults
    - MapDefault
    - MapFormDefault
2. Middleware chaining is a bit awkward &rarr; requires mandatory return, but works.
3. Middleware for logging exists, but the middleware seems to be limited to string based outputs using templates and not fully customizable JSON String messages.
    - However, similar support can be added via `CustomTags` and implementing the JSON log in the same
4. Prometheus Metrics : [#166](https://github.com/gofiber/fiber/issues/166)
5. Route name inside middlewares : [#2195](https://github.com/gofiber/fiber/issues/2195)

## Sources

1. [Fiber Validation](https://copyprogramming.com/howto/how-to-write-a-golang-validation-function-so-that-client-is-able-to-handle-elegantly)
