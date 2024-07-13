# 定义路径
PROTOC_GEN_ECHO_PATH := .\cmd\gen-echo\protoc-gen-echo
PROTOC_GEN_ECHO_SRC_PATH := .\cmd\gen-echo\main.go

PROTOC_GEN_GO_PATH := .\protoc-gen-go.exe
OUT_DIR := .\out

# 声明为 .PHONY 的目标
.PHONY: build-plugin generate-code clean

# 编译插件的目标
build-echo-plugin:
	go build -o $(PROTOC_GEN_ECHO_PATH) ${PROTOC_GEN_ECHO_SRC_PATH}

build-pb:
	protoc -I=protos --plugin=protoc-gen-go=${PROTOC_GEN_GO_PATH} --go_out=protos --go_opt=paths=source_relative ./protos/google/protobuf/descriptor.proto 
	protoc -I=protos --plugin=protoc-gen-go=${PROTOC_GEN_GO_PATH} --go_out=protos --go_opt=paths=source_relative protos/http_options.proto

# 使用插件生成代码的目标
gen-echo-code: build-pb build-echo-plugin
	protoc -I=protos -I example --plugin=protoc-gen-echo=$(PROTOC_GEN_ECHO_PATH) --echo_out=$(OUT_DIR) example/echo/*_handler.proto
	protoc -I=protos -I example --plugin=protoc-gen-go=${PROTOC_GEN_GO_PATH} --go_out=$(OUT_DIR) --go_opt=paths=source_relative example/echo/*_msg.proto

# 清理生成文件的目标
clean:
	rm -f $(PROTOC_GEN_CUSTOM_PATH)
	rm -f $(OUT_DIR)/*.pb.go
	rm -f $(OUT_DIR)/*_echo.go