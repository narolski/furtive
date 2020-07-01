# furtive

*Anonymous Veto Network*-based proof-of-concept for a social game which asks players to respond to sensitive questions. 

The goal of the project was to design and implement a protocol such, that no third party is able to obtain individual players' responses. This is achieved through an AV-net based implementation and a secure, *mTLS*-based method of communication.

*Furtive* client and server applications are written in `go`. The messages between clients and a game server are relayed through a secure websocket connection.