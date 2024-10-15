#!/usr/bin/env bash
set -x
wakeonlan -i 10.0.0.255 f0:de:f1:67:8c:92
wakeonlan -i 10.0.0.255 3c:97:0e:71:77:ab
wakeonlan -i 10.0.0.255 28:d2:44:ed:85:f9
sleep 5