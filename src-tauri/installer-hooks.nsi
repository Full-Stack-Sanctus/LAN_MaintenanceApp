!macro NSIS_HOOK_PREINSTALL
  DetailPrint "Checking for Npcap..."
  IfFileExists "$PROGRAMFILES64\Npcap\Packet.dll" done
  IfFileExists "$PROGRAMFILES\Npcap\Packet.dll" done

  DetailPrint "Installing Npcap..."
  SetOutPath "$PLUGINSDIR"
  
  # Stepping out 5 levels from target/release/bundle/nsis/ to reach src-tauri/drivers/
  File "/oname=npcap-installer.exe" "..\..\..\..\drivers\npcap-installer.exe"
  
  ExecWait '"$PLUGINSDIR\npcap-installer.exe" /winpcap_mode=yes /loopback_support=yes' $0
  DetailPrint "Npcap installation finished with code $0"
done:
!macroend
