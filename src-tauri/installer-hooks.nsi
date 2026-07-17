!include "LogicLib.nsh"

!echo "NSIS script directory = ${__FILEDIR__}"

!macro NSIS_HOOK_PREINSTALL

  DetailPrint "Checking for Npcap..."

  IfFileExists "$SYSDIR\Npcap\Packet.dll" done
  IfFileExists "$PROGRAMFILES\Npcap\Packet.dll" done
  IfFileExists "$PROGRAMFILES64\Npcap\Packet.dll" done

  DetailPrint "Npcap not found."

  SetOutPath "$PLUGINSDIR"

  ; Escapes out of target/release/nsis/x64/ up to the project root folder where drivers/ resides
  File "/oname=npcap-installer.exe" "..\..\..\..\drivers\npcap-installer.exe"
  
  ExecWait '"$PLUGINSDIR\npcap-installer.exe" /S /loopback_support=yes /winpcap_mode=yes' $0

  ${If} $0 != 0
      MessageBox MB_ICONSTOP "Npcap installation failed (exit code $0). Setup will exit."
      Abort
  ${EndIf}

done:

!macroend
