@echo off
setlocal
chcp 65001 >nul
cd /d "%~dp0"
where git >nul 2>nul
if errorlevel 1 (
  echo Please install Git for Windows: https://git-scm.com/download/win
  pause
  exit /b 1
)
set "UPLOAD_TMP=%TEMP%\guo-upload-%RANDOM%-%RANDOM%"
echo Cloning xoman9/guo. Complete browser login if requested.
git clone --branch main https://github.com/xoman9/guo.git "%UPLOAD_TMP%"
if errorlevel 1 goto failed
robocopy "%~dp0." "%UPLOAD_TMP%" /E /XD .git .dart_tool .pub-cache build dist coverage .venv venv __pycache__ .idea /XF .env .env.* *.jks *.keystore *.p12 *.pfx key.properties local.properties >nul
if errorlevel 8 goto failed
pushd "%UPLOAD_TMP%"
git config user.name >nul 2>nul
if errorlevel 1 git config user.name xoman9
git config user.email >nul 2>nul
if errorlevel 1 git config user.email xoman9@users.noreply.github.com
git add -A
if errorlevel 1 goto repo_failed
git diff --cached --quiet
if errorlevel 1 (
  git commit -m "Update guo project"
  if errorlevel 1 goto repo_failed
)
git push origin HEAD:main
if errorlevel 1 goto repo_failed
popd
rmdir /s /q "%UPLOAD_TMP%"
echo Upload completed: https://github.com/xoman9/guo
pause
exit /b 0
:repo_failed
popd
:failed
echo Upload failed. Review the error above. No force push was used.
echo Temporary checkout if created: %UPLOAD_TMP%
pause
exit /b 1
