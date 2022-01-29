## Forked library for [modded TDLib](https://github.com/c0re100/td)

When I'm refactoring my own bot from `Arman92/go-tdlib` to `zelenin/go-tdlib `

I realized that zelenin's library doesn't meet my need😕

So I fork it and make some changes

1. Static build by default
2. Add update event filter
3. Add command parser
4. Receive correct message id to patch text/dice message response.

[Here](example) are a few example codes about how to use **c0re100/gotdlib**.