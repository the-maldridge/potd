FROM docker.io/library/golang:1.26-alpine AS build
RUN apk add --no-cache build-base binutils upx
WORKDIR /build
ARG VERSION="dev"
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 go build -ldflags="-extldflags=-static -X 'github.com/the-maldridge/potd/internal/cmd.Version=$VERSION'" -tags sqlite_omit_load_extension -o potd . && \
    strip potd && \
    upx potd

FROM scratch
COPY --from=build /build/potd potd
USER 1000
ENTRYPOINT ["/potd", "serve"]
