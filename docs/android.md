DonkeyVPN on Android
===

Install [Wireguard application](https://play.google.com/store/apps/details?id=com.wireguard.android&hl=en&pli=1) from Playstore. Add a new tunnel, choose the "Create from scratch" option

![](assets/donkeyvpn-android-14.jpeg)

Generate a keypair, you will need to tap the public key to copy it (it is not very intuitive)

![](assets/donkeyvpn-android-15.jpeg)

Copy the public key and go to the Telegram Bot, and add a new peer using the `/create peer 10.0.0.10/32 <public_key>` command. You can use any ip address in the `10.0.0.0/24` range.

![](assets/donkeyvpn-android-16.jpeg)

Create a new VPN server using `/create vpn` command. Wait until it's ready. Get the information using `/list vpn`.

![](assets/donkeyvpn-android-17.jpeg)

Get server's public key using the `/settings` command and copy it
![](assets/donkeyvpn-android-18.jpeg)

Go back to the Wireguard application and configure you VPN ip address with the same IP used when you executed the `/create peer` command

![](assets/donkeyvpn-android-19.jpeg)

Add a new Peer, use the server's public key you got when executing the `/settings` command. The `Endpoint` field will be the domain name or IP address  plus the `51820` port that you get when you executed the `/list vpn` command, e.g. `12.23.45.67:51820` or `vpn001.example.com:51820`. Set the `Allowed IPs` field with `10.0.0.0/24,0.0.0.0/0` and `Persistent keepalive` with `25`

![](assets/donkeyvpn-android-20.jpeg)

Before connecting, check you actual IP address using [ifconfig.me](https://ifconfig.me/)

![](assets/donkeyvpn-android-21.jpeg)

Connect to the VPN
![](assets/donkeyvpn-android-22.jpeg)

Check back again at [ifconfig.me](https://ifconfig.me/)

![](assets/donkeyvpn-android-23.jpeg)

