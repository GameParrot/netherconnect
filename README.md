# netherconnect
NetherConnect is a proxy that proxies from NetherNet to RakNet, allowing you to join servers with 10-30ms lower ping on windows.

# Downloads
Binaries for Windows can be found in [releases](https://github.com/GameParrot/netherconnect/releases). For other platforms, you can build from source.

# Usage
Sign in with your account (must be the same account that you are signed into Minecraft with), choose your server, and join 127.0.0.1 or "NetherConnect" in LAN games. You will be transferred after you join to connect to the proxy over NetherNet, and your connection will be proxied to the server. If you are on windows, you will be prompted to disable loopback restrictions if you have not already (you will have to accept the UAC dialog that appears.)

# My antivirus says this is a virus!
Golang binaries, especially those that are unsigned and uncommon, can false flag antiviruses. If your antivirus says this is a virus it is a false flag and can be ignored. You can also build from source if you don't want to use the prebuilt binaries

# Server compatibility
This should work on most servers. However, this may not work on some servers, especially ones with anticheats that have strict login checks or ones with excessive ddos protection.
