!macro NSIS_HOOK_PREINSTALL
  ; Tauri flattens/places resource assets inside the TAURI_RESOURCE_DIR root.
  ; We extract the staged npcap installer to a temporary installer directory.
  File "/oname=$PLUGINSDIR\npcap-installer.exe" "${TAURI_RESOURCE_DIR}\npcap-installer.exe"

  ; Run the installer silently with your required arguments
  DetailPrint "Installing Npcap drivers..."
  ExecWait '"$PLUGINSDIR\npcap-installer.exe" /loopback_support=yes /winpcap_mode=yes /S' $0
  
  DetailPrint "Npcap installer exited with code: $0"
!macroend