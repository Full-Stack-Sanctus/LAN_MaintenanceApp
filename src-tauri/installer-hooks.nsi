!macro NSIS_HOOK_PREINSTALL
  ; 'resources' contains files defined in the bundle.resources array
  ; Extract the npcap installer to a temporary directory during setup
  File "/oname=$PLUGINSDIR\npcap-installer.exe" "${TAURI_RESOURCE_DIR}\drivers\npcap-installer.exe"

  ; Run the installer silently with your required arguments
  DetailPrint "Installing Npcap drivers..."
  ExecWait '"$PLUGINSDIR\npcap-installer.exe" /loopback_support=yes /winpcap_mode=yes /S' $0
  
  DetailPrint "Npcap installer exited with code: $0"
!macroend