!macro NSIS_HOOK_PREINSTALL
  # Extracts the installer to the secure temporary staging directory
  File "/oname=$PLUGINSDIR\npcap-installer.exe" "$RESOURCES\drivers\npcap-installer.exe"
  
  # Executes the installer and pauses the main installation thread until it exits
  ExecWait '"$PLUGINSDIR\npcap-installer.exe" /loopback_support=yes /winpcap_mode=yes'
!macroend