!macro NSIS_HOOK_POSTINSTALL
  DetailPrint "Checking system for Npcap requirements..."
  
  ClearErrors
  ReadRegStr $0 HKLM "SOFTWARE\Npcap" ""
  IfErrors 0 NpcapAlreadyInstalled

  DetailPrint "Npcap not found. Launching driver setup..."
  
  IfFileExists "$INSTDIR\drivers\npcap-installer.exe" +3
    MessageBox MB_OK "Error: Npcap installer missing from resources."
    Abort

  ExecWait '"$INSTDIR\resources\drivers\npcap-installer.exe"'

  NpcapAlreadyInstalled:
  DetailPrint "Npcap verification complete."
!macroend
