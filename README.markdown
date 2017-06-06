# Constrained Application Protocol Client and Server for go

You can read more about CoAP in [RFC 7252][coap].  I also did
some preliminary work on `SUBSCRIBE` support from
[an early draft][shelby].

[shelby]: http://tools.ietf.org/html/draft-shelby-core-coap-01
[coap]: http://tools.ietf.org/html/rfc7252

## Differences from original `dustin/go-coap`

1. Added minor helper function to populate URI options.
2. Added configurable receive timeout in backgwards compatible way.  Changed `ResponseTimeout` variable to `DefaultResponseTimeout` to emphasise its role change and to be able to track down its usage within package.
3. Example programs package references have been changed from `dustin/go-coap` to `Kulak/go-coap`.

Motivation for all this change: need to support existing application functionality.