!macro NSIS_HOOK_PREINSTALL
  DetailPrint "Checking for Npcap..."
  IfFileExists "$PROGRAMFILES64\Npcap\Packet.dll" done
  IfFileExists "$PROGRAMFILES\Npcap\Packet.dll" done

  # Look for the driver file exactly where Tauri extracts it
  DetailPrint "Locating bundled Npcap driver..."
  IfFileExists "$INSTDIR\drivers\npcap-installer.exe" run_install
  IfFileExists "$LOCALAPPDATA\lan-maintenanceapp\drivers\npcap-installer.exe" run_install
  
  # Fallback to standard temporary directory extraction if not found yet
  SetOutPath "$PLUGINSDIR"
  File "drivers\npcap-installer.exe"
  Goto run_temp_install

run_install:
  DetailPrint "Installing Npcap from target resource bundle..."
  ExecWait '"$INSTDIR\drivers\npcap-installer.exe" /winpcap_mode=yes /loopback_support=yes' $0
  Goto finish

run_temp_install:
  DetailPrint "Installing Npcap from temporary directory..."
  ExecWait '"$PLUGINSDIR\npcap-installer.exe" /winpcap_mode=yes /loopback_support=yes' $0
  Goto finish

finish:
  DetailPrint "Npcap installation completed with exit code: $0"
done:
!macroend
