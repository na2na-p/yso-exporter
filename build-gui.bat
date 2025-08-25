@echo off
echo Building YSO Exporter GUI...

cd cmd\yso-exporter

echo Generating manifest resource...
rsrc -manifest rsrc.syso.manifest -o rsrc.syso

cd ..\..

echo Building executable with no console window...
go build -ldflags="-H windowsgui" -o yso-exporter.exe ./cmd/yso-exporter

echo Build complete!
echo Run yso-exporter.exe to start the GUI