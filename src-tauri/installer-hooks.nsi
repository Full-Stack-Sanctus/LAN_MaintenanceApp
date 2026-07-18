!include "LogicLib.nsh"

!macro NSIS_HOOK_PREINSTALL

DetailPrint "Checking for Npcap..."

IfFileExists "$PROGRAMFILES64\Npcap\Packet.dll" done
IfFileExists "$PROGRAMFILES\Npcap\Packet.dll" done

DetailPrint "Installing Npcap..."

SetOutPath "$PLUGINSDIR"

File "/oname=npcap-installer.exe" "..\..\..\..\drivers\npcap-installer.exe"

ExecWait '"$PLUGINSDIR\npcap-installer.exe" /winpcap_mode=yes /loopback_support=yes' $0

${If} $0 != 0
    MessageBox MB_ICONSTOP "Npcap installation failed (Exit code $0)"
    Abort
${EndIf}

done:

!macroend