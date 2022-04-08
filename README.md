
# Build

    go build

# Cross compile

    GOOS=windows GOARCH=amd64 go build

# Docu
* https://svn.fhem.de/trac/browser/trunk/fhem/contrib/70_NT5000.pm
* https://wiki.fhem.de/wiki/NT5000

# Features
* commands: getdata, settime
* prometheus interface

# Protocol
## Read online data

Send: "\x00\x01\x02\x01\x04". Last byte is checksum, 5 bytes in total
Receive: 13 bytes in buffer

1. UDC (voltage DC): buffer[0]*2.8+100
2. IDC (current DC): buffer[1]*0.08
3. UAC (voltage AC): buffer[2]+100.0
4. IAC (current AC): buffer[3]*0.120
5. Temperature: buffer[4]-40.0
6. PDC (Power DC): ($udc*$idc)/1000
7. PAC (Power AC): ($uac*$iac)/1000
8. Energy Today: (buffer[6] * 256 + buffer[7])/1000
9. Energy Total: buffer[8] * 256 + buffer[9]; 
