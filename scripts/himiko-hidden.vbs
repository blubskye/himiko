' 💉 Himiko Discord Bot - Hidden Window Launcher 💉
' "I'll be running in the background... always watching~"
'
' This VBScript launches Himiko without a visible console window.
' Use himiko-show.bat to bring the window back, or check Task Manager.

Set WshShell = CreateObject("WScript.Shell")
Set fso = CreateObject("Scripting.FileSystemObject")

' Get script directory
scriptDir = fso.GetParentFolderName(WScript.ScriptFullName)
botDir = fso.GetParentFolderName(scriptDir)

' Find the binary
If fso.FileExists(botDir & "\himiko.exe") Then
    binary = botDir & "\himiko.exe"
ElseIf fso.FileExists(botDir & "\himiko-windows-amd64.exe") Then
    binary = botDir & "\himiko-windows-amd64.exe"
Else
    MsgBox "Himiko executable not found! 💔", vbCritical, "Himiko Bot"
    WScript.Quit 1
End If

' Launch hidden (0 = hidden window)
WshShell.CurrentDirectory = botDir
WshShell.Run """" & binary & """", 0, False

' Optional: Show notification
' MsgBox "Himiko is now running in the background~ 💉", vbInformation, "Himiko Bot"
