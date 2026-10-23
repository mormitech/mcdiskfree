; Inno Setup script for mcdiskfree. Per-user install (no admin rights), adds a Start menu entry, no startup entry.
; Silent install, as winget runs it:  mcdiskfree-setup-<version>-windows-x64.exe /VERYSILENT /SUPPRESSMSGBOXES /NORESTART
; CI replaces __VERSION__ and compiles it (see .github/workflows/build.yml).
[Setup]
AppId={{7D2C4B8E-3A51-4F6D-9C0B-1E8A5F2D7B64}
AppName=mcdiskfree
AppVersion=__VERSION__
AppPublisher=mormitech
AppPublisherURL=https://github.com/mormitech/mcdiskfree
AppSupportURL=https://github.com/mormitech/mcdiskfree/issues
DefaultDirName={autopf}\mcdiskfree
DisableDirPage=yes
DisableProgramGroupPage=yes
DisableReadyPage=yes
PrivilegesRequired=lowest
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
OutputDir=..\..\dist
OutputBaseFilename=mcdiskfree-setup-__VERSION__-windows-x64
UninstallDisplayName=mcdiskfree
UninstallDisplayIcon={app}\mcdiskfree.exe
Compression=lzma2
SolidCompression=yes

[Files]
Source: "..\..\build\bin\mcdiskfree.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\..\LICENSE"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{autoprograms}\mcdiskfree"; Filename: "{app}\mcdiskfree.exe"

[Run]
Filename: "{app}\mcdiskfree.exe"; Description: "Start mcdiskfree"; Flags: nowait postinstall skipifsilent

[Code]
// taskkill returns before Windows has released the file of the killed process, so wait a moment before the installer
// replaces or deletes it.
procedure StopMcdiskfree;
var
  ResultCode: Integer;
begin
  Exec(ExpandConstant('{sys}\taskkill.exe'), '/F /IM mcdiskfree.exe', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Sleep(1500);
end;

// An upgrade replaces a file that may be in use, so close the running app first.
procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssInstall then StopMcdiskfree;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if CurUninstallStep = usUninstall then StopMcdiskfree;
end;
