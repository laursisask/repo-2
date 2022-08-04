# escape=`

FROM mcr.microsoft.com/dotnet/framework/sdk:4.8-20220712-windowsservercore-ltsc2019 as dotnet

SHELL ["powershell", "-Command", "$ErrorActionPreference = 'Stop'; $ProgressPreferences = 'SilentlyContinue';"]

RUN New-Item -ItemType Directory -Path C:\Build; `
    $env:DOTNET_GENERATE_ASPNET_CERTIFICATE | Out-File -FilePath C:\Build\DOTNET_GENERATE_ASPNET_CERTIFICATE; `
    $env:NUGET_VERSION | Out-File -FilePath C:\Build\NUGET_VERSION; `
    $env:ROSLYN_COMPILER_LOCATION | Out-File -FilePath C:\Build\ROSLYN_COMPILER_LOCATION

FROM jenkins/inbound-agent:4.13-2-jdk11-windowsservercore-ltsc2019

COPY --from=dotnet C:/Build C:/Build
COPY --from=dotnet ["/Program Files/dotnet", "/Program Files/dotnet"]
COPY --from=dotnet ["/Program Files/NuGet", "/Program Files/NuGet"]
COPY --from=dotnet ["/Program Files (x86)/Microsoft Visual Studio", "/Program Files (x86)/Microsoft Visual Studio"]
COPY --from=dotnet ["/Program Files (x86)/Microsoft SDKs", "/Program Files (x86)/Microsoft SDKs"]
COPY --from=dotnet ["/Program Files (x86)/Reference Assemblies", "/Program Files (x86)/Reference Assemblies"]

RUN setx PATH $(${Env:PATH} `
    + \";${Env:ProgramFiles}\NuGet\" `
    + \";${Env:ProgramFiles(x86)}\Microsoft Visual Studio\2022\TestAgent\Common7\IDE\CommonExtensions\Microsoft\TestWindow\" `
    + \";${Env:ProgramFiles(x86)}\Microsoft Visual Studio\2022\BuildTools\MSBuild\Current\Bin\" `
    + \";${Env:ProgramFiles(x86)}\Microsoft SDKs\Windows\v10.0A\bin\NETFX 4.8 Tools\" `
    + \";${Env:ProgramFiles(x86)}\Microsoft SDKs\ClickOnce\SignTool\") /m

RUN setx DOTNET_GENERATE_ASPNET_CERTIFICATE $(Get-Content -Path C:\Build\DOTNET_GENERATE_ASPNET_CERTIFICATE) /m; `
    setx NUGET_VERSION $(Get-Content -Path C:\Build\NUGET_VERSION) /m; `
    setx ROSLYN_COMPILER_LOCATION $(Get-Content -Path C:\Build\ROSLYN_COMPILER_LOCATION) /m; `
    Remove-Item -Force -Recurse C:\Build;

# Trigger first run experience by running arbitrary command.
RUN & \"${Env:ProgramFiles}\dotnet\dotnet\" help
