Unicode true

####
## WaterLauncher's installer. It installs for the current user only
## (%LOCALAPPDATA%\Programs\WaterLauncher, no administrator rights), so
## WaterLauncher can update itself later without a UAC prompt.
##
## Build (after `wails3 build`), from this folder:
##   makensis -DARG_WAILS_AMD64_BINARY=..\..\..\bin\WaterLauncher.exe -DINFO_PRODUCTVERSION=1.0.0 project.nsi
## or from the repository root: `wails3 task installer VERSION=v1.0.0`.
##
## Command line (besides NSIS's own /S and /D=<folder>):
##   /relaunch   start WaterLauncher when done (used by its updater)
##   /tray       with /relaunch: start it in the tray
##
## Signing the uninstaller (see docs/SIGNING.md):
##   -DSIGN_CMD="<command>"   signs it while it's built (a local certificate);
##                            the command gets the file path as its last argument.
##   -DINNER                  builds bin\uninstaller-maker.exe, which only writes
##                            bin\uninstall.exe when run, so a remote signer
##                            (SignPath) can sign that file;
##   -DSIGNED_UNINSTALLER=<file>  then packs that signed uninstaller as it is.
####

!define INFO_PROJECTNAME    "WaterLauncher"
!define INFO_COMPANYNAME    "ApolloF"
!define INFO_PRODUCTNAME    "WaterLauncher"
!ifndef INFO_PRODUCTVERSION
    !define INFO_PRODUCTVERSION "0.0.0"
!endif
!define INFO_COPYRIGHT      "(c) 2026 ApolloF, MIT License"
!define PRODUCT_EXECUTABLE  "WaterLauncher.exe"
!define UNINST_KEY_NAME     "WaterLauncher"
!define WAILS_INSTALL_SCOPE "user"
!define REQUEST_EXECUTION_LEVEL "user"
!define RUN_KEY "Software\Microsoft\Windows\CurrentVersion\Run"
!define STARTUP_APPROVED_KEY "Software\Microsoft\Windows\CurrentVersion\Explorer\StartupApproved\Run"

!include "wails_tools.nsh"
!include "LogicLib.nsh"
!include "FileFunc.nsh"

# The version information needs four parts.
VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"

VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Setup"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

ManifestDPIAware true
SetCompressor /SOLID lzma

!ifdef SIGN_CMD
    !uninstfinalize '${SIGN_CMD} "%1"' = 0
!endif

!include "MUI2.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
!define MUI_ABORTWARNING
!define MUI_FINISHPAGE_RUN "$INSTDIR\${PRODUCT_EXECUTABLE}"
!define MUI_FINISHPAGE_RUN_TEXT "Start WaterLauncher"

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

Name "${INFO_PRODUCTNAME}"
!ifdef INNER
    OutFile "..\..\..\bin\uninstaller-maker.exe"
!else
    OutFile "..\..\..\bin\WaterLauncher-setup.exe"
!endif
InstallDir "$LOCALAPPDATA\Programs\${INFO_PRODUCTNAME}"
# An update or reinstall goes where WaterLauncher already is.
InstallDirRegKey HKCU "${UNINST_KEY}" "InstallLocation"
ShowInstDetails nevershow
ShowUninstDetails nevershow

# Closes a running WaterLauncher (it may sit in the tray) and waits until
# its program file can be replaced. $0 is the folder to look in.
!macro CloseWaterLauncher UN
Function ${UN}CloseWaterLauncher
    IfFileExists "$0\${PRODUCT_EXECUTABLE}" 0 done
    # "--quit" asks the running copy to close; it exits at once when none runs.
    ExecWait '"$0\${PRODUCT_EXECUTABLE}" --quit'
    StrCpy $1 0
    retry:
        ClearErrors
        # Opening the exe for writing fails while it runs.
        FileOpen $2 "$0\${PRODUCT_EXECUTABLE}" a
        IfErrors 0 closed
        IntOp $1 $1 + 1
        IntCmp $1 60 0 wait 0
            IfSilent fail
            MessageBox MB_RETRYCANCEL|MB_ICONEXCLAMATION "WaterLauncher is still running. Close it (also from the tray), then choose Retry." IDRETRY again
            fail:
            SetErrorLevel 2
            Quit
            again:
            StrCpy $1 0
        wait:
        Sleep 250
        Goto retry
    closed:
        FileClose $2
    done:
FunctionEnd
!macroend
!insertmacro CloseWaterLauncher ""
!insertmacro CloseWaterLauncher "un."

Function .onInit
!ifdef INNER
    # Only write the uninstaller next to this exe, for signing.
    WriteUninstaller "$EXEDIR\uninstall.exe"
    Quit
!endif
    !insertmacro wails.checkArchitecture
FunctionEnd

Section
    !insertmacro wails.setShellContext

    StrCpy $0 $INSTDIR
    Call CloseWaterLauncher

    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR
    !insertmacro wails.files
    # Left behind by an update of a copy that wasn't installed.
    Delete "$INSTDIR\${PRODUCT_EXECUTABLE}.old"

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    # The desktop shortcut comes with a first install; updates don't bring
    # back one you deleted.
    IfSilent +2
        CreateShortcut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

!ifdef SIGNED_UNINSTALLER
    File "/oname=uninstall.exe" "${SIGNED_UNINSTALLER}"
!else
    WriteUninstaller "$INSTDIR\uninstall.exe"
!endif
    SetRegView 64
    WriteRegStr HKCU "${UNINST_KEY}" "Publisher" "${INFO_COMPANYNAME}"
    WriteRegStr HKCU "${UNINST_KEY}" "DisplayName" "${INFO_PRODUCTNAME}"
    WriteRegStr HKCU "${UNINST_KEY}" "DisplayVersion" "${INFO_PRODUCTVERSION}"
    WriteRegStr HKCU "${UNINST_KEY}" "DisplayIcon" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    WriteRegStr HKCU "${UNINST_KEY}" "UninstallString" "$\"$INSTDIR\uninstall.exe$\""
    WriteRegStr HKCU "${UNINST_KEY}" "QuietUninstallString" "$\"$INSTDIR\uninstall.exe$\" /S"
    ${GetSize} "$INSTDIR" "/S=0K" $0 $1 $2
    IntFmt $0 "0x%08X" $0
    WriteRegDWORD HKCU "${UNINST_KEY}" "EstimatedSize" "$0"
    WriteRegStr HKCU "${UNINST_KEY}" "InstallLocation" "$INSTDIR"
    WriteRegStr HKCU "${UNINST_KEY}" "URLInfoAbout" "https://github.com/ApolloF/WaterLauncher"
    WriteRegDWORD HKCU "${UNINST_KEY}" "NoModify" 1
    WriteRegDWORD HKCU "${UNINST_KEY}" "NoRepair" 1

    # Started by WaterLauncher's updater: start it again.
    ${GetParameters} $R0
    ClearErrors
    ${GetOptions} $R0 "/relaunch" $R1
    IfErrors noRelaunch
        ClearErrors
        ${GetOptions} $R0 "/tray" $R1
        IfErrors 0 +3
            Exec '"$INSTDIR\${PRODUCT_EXECUTABLE}" --updated'
            Goto noRelaunch
        Exec '"$INSTDIR\${PRODUCT_EXECUTABLE}" --updated --tray'
    noRelaunch:
SectionEnd

Section "uninstall"
    !insertmacro wails.setShellContext

    StrCpy $0 $INSTDIR
    Call un.CloseWaterLauncher

    # Only WaterLauncher's own files: the folder may have been chosen by hand.
    Delete "$INSTDIR\${PRODUCT_EXECUTABLE}"
    Delete "$INSTDIR\${PRODUCT_EXECUTABLE}.old"
    Delete "$INSTDIR\${PRODUCT_EXECUTABLE}.new"
    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    # Starting with Windows.
    DeleteRegValue HKCU "${RUN_KEY}" "${INFO_PRODUCTNAME}"
    DeleteRegValue HKCU "${STARTUP_APPROVED_KEY}" "${INFO_PRODUCTNAME}"

    IfSilent keepData
    MessageBox MB_YESNO|MB_ICONQUESTION|MB_DEFBUTTON2 "Also delete your WaterLauncher library, settings, saved keys and downloaded art?$\r$\n$\r$\nYour games and their saves are not touched either way." IDNO keepData
        RMDir /r "$APPDATA\WaterLauncher"
        RMDir /r "$LOCALAPPDATA\WaterLauncher"
    keepData:

    !insertmacro wails.deleteUninstaller
    RMDir $INSTDIR
SectionEnd
