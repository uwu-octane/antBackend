# 构建阶段
FROM postgres:17-alpine AS builder
# + 安装 llvm19 / clang19
RUN apk add --no-cache build-base git postgresql-dev llvm19 clang19
WORKDIR /build
RUN git clone https://github.com/andrielfn/pg-ulid.git
WORKDIR /build/pg-ulid
RUN make && make install

# 运行阶段
FROM postgres:17-alpine
COPY --from=builder /usr/local/lib/postgresql/ulid.so /usr/local/lib/postgresql/
COPY --from=builder /usr/local/share/postgresql/extension/ulid* /usr/local/share/postgresql/extension/