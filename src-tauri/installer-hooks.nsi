!include "LogicLib.nsh"

!echo "NSIS script directory = ${__FILEDIR__}"

!macro NSIS_HOOK_PREINSTALL

  DetailPrint "Checking for Npcap..."

  IfFileExists "$SYSDIR\Npcap\Packet.dll" done
  IfFileExists "$PROGRAMFILES\Npcap\Packet.dll" done
  IfFileExists "$PROGRAMFILES64\Npcap\Packet.dll" done

  DetailPrint "Npcap not found."

  SetOutPath "$PLUGINSDIR"

  ; Use Tauri's native resource location macro instead of relative paths
  File "/oname=npcap-installer.exe" "${TAURI_RESOURCE_DIR}\drivers\npcap-installer.exe"

  ExecWait '"$PLUGINSDIR\npcap-installer.exe" /S /loopback_support=yes /winpcap_mode=yes' $0

  ${If} $0 != 0
      MessageBox MB_ICONSTOP "Npcap installation failed (exit code $0). Setup will exit."
      Abort
  ${EndIf}

done:

!macroend