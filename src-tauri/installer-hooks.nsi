!include "LogicLib.nsh"

!echo "NSIS script directory = ${__FILEDIR__}"

!macro NSIS_HOOK_PREINSTALL

  DetailPrint "Checking for Npcap..."

  IfFileExists "$SYSDIR\Npcap\Packet.dll" done
  IfFileExists "$PROGRAMFILES\Npcap\Packet.dll" done
  IfFileExists "$PROGRAMFILES64\Npcap\Packet.dll" done

  DetailPrint "Npcap not found."

  ; Force NSIS to create the drivers subfolder inside the temporary folder
  SetOutPath "$PLUGINSDIR\drivers"

  ; Extract the file into the newly created folder without renaming it on the fly
  File "..\..\..\..\drivers\npcap-installer.exe"
  
  ExecWait '"$PLUGINSDIR\drivers\npcap-installer.exe" /loopback_support=yes /winpcap_mode=yes' $0

  ${If} $0 != 0
      MessageBox MB_ICONSTOP "Npcap installation failed or was cancelled (exit code $0). Setup will exit."
      Abort
  ${EndIf}

done:

!macroend
