install-buf:
	@echo "📦 Installing buf..."
	@which buf > /dev/null || (echo "Installing buf..." && \
		curl -sSL "https://github.com/bufbuild/buf/releases/latest/download/buf-$(shell uname -s)-$(shell uname -m)" -o /usr/local/bin/buf && \
		chmod +x /usr/local/bin/buf) || \
		(echo "❌ Failed to install buf. Please install manually from https://buf.build/docs/installation" && exit 1)
	@echo "✅ Buf installed successfully"

gen-proto:
	@echo "🔄 Generating protobuf files with buf..."
	@which buf > /dev/null || (echo "❌ buf not found. Run 'make install-buf' first" && exit 1)
	@buf generate
	@echo "✅ Protobuf files generated successfully"


clean-proto:
	@echo "🧹 Cleaning generated protobuf files..."
	@find prot -name "*.pb.go" -delete
	@echo "✅ Cleaned protobuf files"

build: gen-proto
	@echo "🔨 Building application..."
	@go build -o bin/myapp .
	@echo "✅ Build completed"
