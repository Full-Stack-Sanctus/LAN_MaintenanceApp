!include "LogicLib.nsh"

!echo "NSIS script directory = ${__FILEDIR__}"

!macro NSIS_HOOK_PREINSTALL

  DetailPrint "Checking for Npcap..."

  IfFileExists "$SYSDIR\Npcap\Packet.dll" done
  IfFileExists "$PROGRAMFILES\Npcap\Packet.dll" done
  IfFileExists "$PROGRAMFILES64\Npcap\Packet.dll" done

  DetailPrint "Npcap not found."

  SetOutPath "$PLUGINSDIR"

  File "/oname=npcap-installer.exe" "..\..\..\..\drivers\npcap-installer.exe"
  
  ; REMOVED THE /S FLAG: Allows the regular interactive window to load for the user
  ExecWait '"$PLUGINSDIR\npcap-installer.exe" /loopback_support=yes /winpcap_mode=yes' $0

  ${If} $0 != 0
      MessageBox MB_ICONSTOP "Npcap installation failed or was cancelled (exit code $0). Setup will exit."
      Abort
  ${EndIf}

done:

!macroend
