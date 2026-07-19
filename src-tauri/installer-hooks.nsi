!macro NSIS_HOOK_PREINSTALL
  DetailPrint "Checking for Npcap..."
  IfFileExists "$PROGRAMFILES64\Npcap\Packet.dll" done
  IfFileExists "$PROGRAMFILES\Npcap\Packet.dll" done

  DetailPrint "Installing Npcap..."
  
  # 1. Instruct NSIS to use a flat temporary workspace path
  SetOutPath "$PLUGINSDIR"
  
  # 2. Use Tauri's built-in PROJECT_DIR macro to safely find your repository root files
  File "${PROJECT_DIR}\drivers\npcap-installer.exe"
  
  # 3. Fire the installer thread and wait for completion safely
  ExecWait '"$PLUGINSDIR\npcap-installer.exe" /winpcap_mode=yes /loopback_support=yes' $0
  DetailPrint "Npcap installation finished with code $0"
done:
!macroend
