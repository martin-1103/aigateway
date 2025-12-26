# Streaming Quick Reference

## Files Created

### Core Streaming Infrastructure
```
D:\temp\aigateway\providers\stream\forwarder.go       (70 lines)
```

### OpenAI Provider
```
D:\temp\aigateway\providers\openai\stream.executor.go    (127 lines)
D:\temp\aigateway\providers\openai\stream.translator.go  (67 lines)
```

### GLM Provider
```
D:\temp\aigateway\providers\glm\stream.executor.go       (132 lines)
D:\temp\aigateway\providers\glm\stream.translator.go     (67 lines)
```

### Antigravity Provider
```
D:\temp\aigateway\providers\antigravity\stream.adapter.go     (75 lines)
D:\temp\aigateway\providers\antigravity\stream.translator.go  (69 lines)
```

## Files Modified

```
D:\temp\aigateway\providers\provider.go               (+24 lines) - Added interface methods
D:\temp\aigateway\providers\openai\provider.go        (+50 lines) - Implemented streaming
D:\temp\aigateway\providers\glm\provider.go           (+31 lines) - Implemented streaming
D:\temp\aigateway\providers\antigravity\provider.go   (+50 lines) - Implemented streaming
D:\temp\aigateway\services\executor.service.go        (+67 lines) - Added ExecuteStream()
D:\temp\aigateway\handlers\proxy.handler.go           (+90 lines) - Added streaming handler
```

## Key Components

### 1. StreamResponse Structure
```go
type StreamResponse struct {
    StatusCode int
    Headers    map[string]string
    DataCh     <-chan []byte      // SSE chunks
    ErrCh      <-chan error       // Errors
    Done       <-chan struct{}    // Completion signal
}
```

### 2. Provider Interface Methods
```go
ExecuteStream(ctx context.Context, req *ExecuteRequest) (*StreamResponse, error)
SupportsStreaming() bool
```

### 3. Translation Functions

**OpenAI:**
```go
TranslateOpenAIStreamToClaude(chunk []byte) []byte
```

**GLM:**
```go
TranslateGLMStreamToClaude(chunk []byte) []byte
```

**Antigravity:**
```go
TranslateAntigravityStreamToClaude(data []byte, eventType string) []byte
```

## Request Flow

```
1. Client sends request with "stream": true
   ↓
2. ProxyHandler.HandleProxy() detects stream flag
   ↓
3. ProxyHandler.handleStreaming() sets SSE headers
   ↓
4. ExecutorService.ExecuteStream() orchestrates pipeline
   ↓
5. Provider.ExecuteStream() makes HTTP request
   ↓
6. Stream executor reads SSE chunks
   ↓
7. Translator converts to Claude format
   ↓
8. Chunks sent to DataCh channel
   ↓
9. Handler reads from DataCh and writes to client
   ↓
10. Handler flushes after each chunk
```

## SSE Event Types

### message_start
```
event: message_start
data: {"type":"message_start","message":{"role":"assistant"}}
```

### content_block_delta
```
event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}
```

### message_stop
```
event: message_stop
data: {"type":"message_stop"}
```

### error (on failure)
```
event: error
data: {"error": "error message"}
```

## Testing Commands

### Test OpenAI Streaming
```bash
curl -N -X POST http://localhost:8088/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "stream": true,
    "messages": [{"role": "user", "content": "Count to 5"}]
  }'
```

### Test GLM Streaming
```bash
curl -N -X POST http://localhost:8088/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "glm-4",
    "stream": true,
    "messages": [{"role": "user", "content": "数到5"}]
  }'
```

### Test Antigravity Streaming
```bash
curl -N -X POST http://localhost:8088/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.0-flash-exp",
    "stream": true,
    "messages": [{"role": "user", "content": "Count to 5"}]
  }'
```

**Note:** The `-N` flag disables curl's buffering for immediate output.

## Implementation Patterns

### Channel Pattern
```go
dataCh := make(chan []byte, 10)
errCh := make(chan error, 1)
done := make(chan struct{})

go func() {
    defer close(dataCh)
    defer close(errCh)
    defer close(done)
    // ... read stream
}()
```

### SSE Reading Pattern (OpenAI/GLM)
```go
scanner := bufio.NewScanner(body)
for scanner.Scan() {
    line := scanner.Bytes()
    if bytes.HasPrefix(line, []byte("data: ")) {
        data := line[6:]
        if bytes.Equal(data, []byte("[DONE]")) {
            break
        }
        dataCh <- data
    }
}
```

### SSE Writing Pattern (Handler)
```go
for {
    select {
    case data, ok := <-streamResp.DataCh:
        if !ok { return }
        c.Writer.Write(data)
        flusher.Flush()
    case err := <-streamResp.ErrCh:
        // Handle error
    case <-streamResp.Done:
        return
    case <-c.Request.Context().Done():
        return
    }
}
```

## Common Issues & Solutions

### Issue: Stream not starting
**Solution:** Check if provider.SupportsStreaming() returns true

### Issue: Buffering delay
**Solution:** Ensure flusher.Flush() is called after each write

### Issue: Memory leak
**Solution:** All channels must be closed in defer statements

### Issue: Chunk format error
**Solution:** Verify translator outputs valid SSE format with "event:" and "data:" lines

### Issue: Stats not recorded
**Solution:** Stats are recorded after stream completion via goroutine

## Performance Benchmarks

Expected performance:
- **First chunk latency:** < 500ms
- **Inter-chunk latency:** < 100ms
- **Memory per stream:** ~100KB (buffered channels)
- **Concurrent streams:** Limited by provider rate limits

## Code Quality Checklist

- ✅ All files under 300 lines
- ✅ Clear separation of concerns
- ✅ Proper error handling at boundaries
- ✅ Channel cleanup via defer
- ✅ Context cancellation support
- ✅ No data races (chunk copying)
- ✅ Graceful shutdown on errors
- ✅ Consistent SSE format across providers

## Next Steps

1. **Test with real providers** - Requires valid API keys
2. **Monitor memory usage** - Check for leaks during long streams
3. **Add metrics** - Track streaming performance
4. **Load testing** - Verify concurrent streaming capacity
5. **Integration tests** - Automated end-to-end tests
