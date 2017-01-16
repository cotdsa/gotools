### Windows notes

Add the address on the loopback interface with
```
netsh interface ipv4 add address "Loopback Pseudo-Interface 1" 169.254.169.254 255.255.255.255
```
This setting will persist across reboots

### OSX notes
Add the address on the loopback interface with
```
sudo ifconfig lo0 alias 169.254.169.254
```
This setting will *NOT* persist across reboots
