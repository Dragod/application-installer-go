; Package Installer - Inno Setup Script
; User-mode installer for Package Installer application

[Setup]
; Application information
AppName=PfCode - Application Installer
AppVersion=1.0.0
AppVerName=PfCode - Application Installer 1.0.0
AppPublisher=PfCode
AppPublisherURL=https://github.com/user/pf-installer
AppSupportURL=https://github.com/user/pf-installer
AppUpdatesURL=https://github.com/user/pf-installer
AppCopyright=Copyright (C) 2025 PfCode

; Installation directories (user-mode installation)
DefaultDirName={userpf}\PfCode - Application Installer
DefaultGroupName=PfCode - Application Installer
AllowNoIcons=yes

; Output
OutputDir=.
OutputBaseFilename=PfCode-ApplicationInstaller-Setup
Compression=lzma
SolidCompression=yes

; User-mode installation (no admin required)
PrivilegesRequired=lowest
PrivilegesRequiredOverridesAllowed=commandline

; Misc
WizardStyle=modern
DisableProgramGroupPage=yes
LicenseFile=LICENSE.txt
InfoBeforeFile=README.md

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[Files]
Source: "pfcode-installer.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "README.md"; DestDir: "{app}"; Flags: ignoreversion
Source: "favicon.ico"; DestDir: "{app}"; Flags: ignoreversion
; NOTE: Don't use "Flags: ignoreversion" on any shared system files

[Icons]
Name: "{group}\PfCode - Application Installer"; Filename: "{app}\pfcode-installer.exe"; WorkingDir: "{app}"; IconFilename: "{app}\favicon.ico"
Name: "{group}\{cm:UninstallProgram,PfCode - Application Installer}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\PfCode - Application Installer"; Filename: "{app}\pfcode-installer.exe"; WorkingDir: "{app}"; IconFilename: "{app}\favicon.ico"; Tasks: desktopicon

[Registry]
; File association for CSV import (user-mode)
Root: HKCU; Subkey: "SOFTWARE\Classes\.csv\OpenWithProgids"; ValueType: string; ValueName: "PfCodeInstaller.csv"; ValueData: ""; Flags: uninsdeletevalue
Root: HKCU; Subkey: "SOFTWARE\Classes\PfCodeInstaller.csv"; ValueType: string; ValueName: ""; ValueData: "PfCode Application Installer List"; Flags: uninsdeletekey
Root: HKCU; Subkey: "SOFTWARE\Classes\PfCodeInstaller.csv\shell\Import with PfCode"; ValueType: string; ValueName: ""; ValueData: "Import with PfCode - Application Installer"; Flags: uninsdeletekey
Root: HKCU; Subkey: "SOFTWARE\Classes\PfCodeInstaller.csv\shell\Import with PfCode\command"; ValueType: string; ValueName: ""; ValueData: """{app}\pfcode-installer.exe"" ""%1"""; Flags: uninsdeletekey

[Run]
Filename: "{app}\pfcode-installer.exe"; Description: "{cm:LaunchProgram,PfCode - Application Installer}"; Flags: nowait postinstall skipifsilent; WorkingDir: "{app}"

[Dirs]
Name: "{userappdata}\PfCode - Application Installer"; Flags: uninsneveruninstall

[Code]
procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssInstall then
  begin
    // Create application data directory
    ForceDirectories(ExpandConstant('{userappdata}\PfCode - Application Installer'));
  end;
end; 