Section "Install Npcap" SecNpcap
  ; 1. Check if Npcap is already installed by looking at the registry
  ClearErrors
  ReadRegStr $0 HKLM "SOFTWARE\Npcap" ""
  
  ; If the registry key exists, skip to the end
  IfErrors 0 NpcapAlreadyInstalled

  ; 2. If it does not exist, extract and run your installer
  SetOutPath "$pluginsdir"
  File "src-tauri\drivers\npcap-installer.exe"
  
  DetailPrint "Installing Npcap Driver..."
  
  ; For FREE version: Runs the interactive installer
  ExecWait '"$pluginsdir\npcap-installer.exe"'
  
  ; For OEM version (Uncomment below and comment out the line above if you buy OEM):
  ; ExecWait '"$pluginsdir\npcap-installer.exe" /S'

  NpcapAlreadyInstalled:
SectionEnd
