@echo on

@rem 删除GIT版本控制目录

@rem for /r . %%a in (.) do @if exist "%%a\.svn" @echo "%%a\.git"
@for /r . %%a in (.) do @if exist "%%a\.git" rd /s /q "%%a\.git"

@echo completed
@pause