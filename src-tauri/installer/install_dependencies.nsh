File "/oname=$PLUGINSDIR\npcap-installer.exe" "${NSISDIR}\..\drivers\npcap-installer.exe"
ExecWait '"$PLUGINSDIR\npcap-installer.exe" /loopback_support=yes /winpcap_mode=yes'