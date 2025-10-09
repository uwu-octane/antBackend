#!/usr/bin/env bash
set -euo pipefail

usage() { echo "用法: $0 {proto|pb|all} <module>"; exit 1; }
[[ $# -ge 2 ]] || usage

CMD="$1"
MODNAME="$2"                               # 例如 test
ROOTMOD="$(go list -m -f '{{.Path}}')"     # 仓库根模块，比如 github.com/taoting/antBackend
PKG="api/api/v1/$MODNAME"
PROTO="$PKG/$MODNAME.proto"
CAP="$(printf "%s" "$MODNAME" | awk '{print toupper(substr($0,1,1)) substr($0,2)}')"

ensure_proto () {
  mkdir -p "$PKG"
  if [[ ! -f "$PROTO" ]]; then
    cat > "$PROTO" <<EOF
syntax = "proto3";

package $MODNAME.v1;
option go_package = "$ROOTMOD/$PKG;${MODNAME}pb";

service ${CAP}Service {
  rpc Ping(PingReq) returns (PingResp);
}

message PingReq {}
message PingResp {}
EOF
    echo "已创建 proto: $PROTO"
  fi
}

gen_pb () {
  protoc -I . "$PROTO" \
    --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative
}

gen_zrpc () {
  goctl rpc protoc "$PROTO" \
    --zrpc_out="apps/$MODNAME" \
    --module="$ROOTMOD"
}

case "$CMD" in
  proto) ensure_proto ;;
  pb)    ensure_proto; gen_pb ;;
  all)   ensure_proto; gen_pb; gen_zrpc ;;
  *)     usage ;;
esac