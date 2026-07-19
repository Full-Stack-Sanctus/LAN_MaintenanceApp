!macro NSIS_HOOK_PREINSTALL
  DetailPrint "Checking for Npcap..."
  IfFileExists "$PROGRAMFILES64\Npcap\Packet.dll" done
  IfFileExists "$PROGRAMFILES\Npcap\Packet.dll" done

  DetailPrint "Installing Npcap..."
  
  # 1. Set the clean, flat temporary extraction directory
  SetOutPath "$PLUGINSDIR"
  
  # 2. Extract the file directly into $PLUGINSDIR without naming subfolders
  File "..\..\..\..\drivers\npcap-installer.exe"
  
  # 3. Execute the executable straight from the root plugin workspace path
  ExecWait '"$PLUGINSDIR\npcap-installer.exe" /winpcap_mode=yes /loopback_support=yes' $0
  DetailPrint "Npcap installation finished with code $0"
done:
!macroend
