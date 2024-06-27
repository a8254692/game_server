@echo off
@echo Compiling...
rem 获取当前batch(build.bat)的路径
@set cue_path=%~dp0

if not exist gmsg (
	md gmsg
)

protoc --proto_path=.\common_proto --go_out=%cue_path%\gmsg\ --plugin=protoc-gen-go=.\protoc-gen-go.exe  .\common_proto\*.proto
protoc --proto_path=.\inner_proto --go_out=%cue_path%\gmsg\ --plugin=protoc-gen-go=.\protoc-gen-go.exe  .\inner_proto\*.proto
@echo compile success

@pause