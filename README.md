# go-smallchat
This is an example of  chat server program inspired by the [`smallchat`](https://github.com/antirez/smallchat) project by antirez. For learning and practice, I wrote a version in Go language.

 This project serves as a learning tool designed to understand the basics of TCP server programming and how a simple chat server can be constructed and operated.


Like smallchat, you can connect to the chat server using telnet or netcat on the default port `8080`:


```

$ telnet localhost 8080

```
or
```

$ nc localhost 8080

```

### Commands
- `/nick <name>` - Set your nickname in the chat.
- `/quit` - Disconnect from the chat server.

### Disclaimer
This project is for educational purposes only. It is not intended for production use as it lacks comprehensive error handling and security features.