package parser_test

import (
	"net/netip"
	"testing"
	"time"

	"github.com/HT4w5/nyaago/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNginxJSONParser_Parse(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		line := []byte(`{"timestamp":1734345934.123,"remote_addr":"127.0.0.1","request_method":"GET","request_uri":"/index.html","status":"200","body_bytes_sent":1024}`)
		parser := &parser.NginxJSONParser{}

		req, err := parser.Parse(line)
		require.NoError(t, err)

		assert.Equal(t, "/index.html", req.URL)
		assert.Equal(t, int64(1024), req.BodySent)
		assert.Equal(t, time.Unix(1734345934, 0), req.Time)
		assert.Equal(t, netip.MustParseAddr("127.0.0.1"), req.Client)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		line := []byte(`{"timestamp":1734345934.123,"remote_addr":"127.0.0.1","request_method":"GET","request_uri":"/index.html","status":"200","body_bytes_sent":1024`)
		parser := &parser.NginxJSONParser{}

		_, err := parser.Parse(line)
		require.Error(t, err)
	})

	t.Run("Missing remote_addr", func(t *testing.T) {
		line := []byte(`{"timestamp":1734345934.123,"request_method":"GET","request_uri":"/index.html","status":"200","body_bytes_sent":1024}`)
		parser := &parser.NginxJSONParser{}

		_, err := parser.Parse(line)
		require.Error(t, err)
	})

	t.Run("Invalid remote_addr", func(t *testing.T) {
		line := []byte(`{"timestamp":1734345934.123,"remote_addr":"invalid","request_method":"GET","request_uri":"/index.html","status":"200","body_bytes_sent":1024}`)
		parser := &parser.NginxJSONParser{}

		_, err := parser.Parse(line)
		require.Error(t, err)
	})

	t.Run("Negative body_bytes_sent", func(t *testing.T) {
		line := []byte(`{"timestamp":1734345934.123,"remote_addr":"127.0.0.1","request_method":"GET","request_uri":"/index.html","status":"200","body_bytes_sent":-1024}`)
		parser := &parser.NginxJSONParser{}

		req, err := parser.Parse(line)
		require.NoError(t, err)

		assert.Equal(t, int64(-1024), req.BodySent)
	})

	t.Run("Empty request_uri", func(t *testing.T) {
		line := []byte(`{"timestamp":1734345934.123,"remote_addr":"127.0.0.1","request_method":"GET","request_uri":"","status":"200","body_bytes_sent":1024}`)
		parser := &parser.NginxJSONParser{}

		req, err := parser.Parse(line)
		require.NoError(t, err)

		assert.Equal(t, "", req.URL)
	})

	t.Run("IPv6 remote_addr", func(t *testing.T) {
		line := []byte(`{"timestamp":1734345934.123,"remote_addr":"::1","request_method":"GET","request_uri":"/index.html","status":"200","body_bytes_sent":1024}`)
		parser := &parser.NginxJSONParser{}

		req, err := parser.Parse(line)
		require.NoError(t, err)

		assert.Equal(t, netip.MustParseAddr("::1"), req.Client)
	})

	t.Run("Large timestamp", func(t *testing.T) {
		line := []byte(`{"timestamp":1834345934.123,"remote_addr":"127.0.0.1","request_method":"GET","request_uri":"/index.html","status":"200","body_bytes_sent":1024}`)
		parser := &parser.NginxJSONParser{}

		req, err := parser.Parse(line)
		require.NoError(t, err)

		assert.Equal(t, time.Unix(1834345934, 0), req.Time)
	})

	t.Run("Small timestamp", func(t *testing.T) {
		line := []byte(`{"timestamp":1634345934.123,"remote_addr":"127.0.0.1","request_method":"GET","request_uri":"/index.html","status":"200","body_bytes_sent":1024}`)
		parser := &parser.NginxJSONParser{}

		req, err := parser.Parse(line)
		require.NoError(t, err)

		assert.Equal(t, time.Unix(1634345934, 0), req.Time)
	})

	t.Run("Zero timestamp", func(t *testing.T) {
		line := []byte(`{"timestamp":0,"remote_addr":"127.0.0.1","request_method":"GET","request_uri":"/index.html","status":"200","body_bytes_sent":1024}`)
		parser := &parser.NginxJSONParser{}

		req, err := parser.Parse(line)
		require.NoError(t, err)

		assert.Equal(t, time.Unix(0, 0), req.Time)
	})

	t.Run("Negative timestamp", func(t *testing.T) {
		line := []byte(`{"timestamp":-1734345934.123,"remote_addr":"127.0.0.1","request_method":"GET","request_uri":"/index.html","status":"200","body_bytes_sent":1024}`)
		parser := &parser.NginxJSONParser{}

		req, err := parser.Parse(line)
		require.NoError(t, err)

		assert.Equal(t, time.Unix(-1734345934, 0), req.Time)
	})

	t.Run("Zero body_bytes_sent", func(t *testing.T) {
		line := []byte(`{"timestamp":1734345934.123,"remote_addr":"127.0.0.1","request_method":"GET","request_uri":"/index.html","status":"200","body_bytes_sent":0}`)
		parser := &parser.NginxJSONParser{}

		req, err := parser.Parse(line)
		require.NoError(t, err)

		assert.Equal(t, int64(0), req.BodySent)
	})
}
