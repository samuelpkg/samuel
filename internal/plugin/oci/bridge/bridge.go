// Package bridge implements the host-side server an OCI plugin
// container talks to over /samuel-bridge (a Unix-domain socket bind-
// mounted into the container's filesystem).
//
// The wire protocol matches api/proto/plugin/v1/plugin.proto. v2.0
// ships JSON-over-Unix-socket as the primary transport (one request
// per connection, length-prefixed by the HTTP request line) so we can
// land end-to-end tests today without dragging in the protoc toolchain
// for v2.0-beta.1. Generated gRPC bindings ship in v2.1 alongside the
// first real OCI plugin (Milestone 5: claude-runner).
//
// Server lifecycle:
//
//	srv := bridge.NewServer(socketPath)
//	srv.HandleHook("pre-commit", func(req *Request) (*Response, error) { ... })
//	go srv.Serve()
//	... (container does its work, makes RPCs)
//	srv.Stop()
//
// The Server is safe for concurrent use.
package bridge

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
)

// Component is the structured-error namespace.
const Component = "plugin/oci/bridge"

// Method names mirror the proto RPCs.
const (
	MethodDetect    = "PluginService/Detect"
	MethodInstall   = "PluginService/Install"
	MethodCheck     = "PluginService/Check"
	MethodUninstall = "PluginService/Uninstall"
	MethodHook      = "PluginService/Hook"
)

// Capability is the wire form of api.Capability.
type Capability struct {
	Kind    string   `json:"kind"`
	Targets []string `json:"targets,omitempty"`
}

// Mutation mirrors api.Mutation (Reverse closures don't cross the
// wire — only kind/path/description).
type Mutation struct {
	Kind        string `json:"kind"`
	Path        string `json:"path"`
	Description string `json:"description,omitempty"`
}

// Request is the unified wire envelope. Method names match the const
// values above. Body is the method-specific payload.
type Request struct {
	Method string          `json:"method"`
	Body   json.RawMessage `json:"body,omitempty"`
}

// Response is the unified wire envelope.
type Response struct {
	OK    bool            `json:"ok"`
	Error string          `json:"error,omitempty"`
	Body  json.RawMessage `json:"body,omitempty"`
}

// Per-method request/response types match the proto messages. They are
// defined here to keep the bridge self-contained.

type DetectRequest struct{}

type DetectResponse struct {
	Installed bool   `json:"installed"`
	Version   string `json:"version,omitempty"`
	Path      string `json:"path,omitempty"`
}

type InstallRequest struct {
	DryRun  bool         `json:"dry_run,omitempty"`
	Force   bool         `json:"force,omitempty"`
	Verbose bool         `json:"verbose,omitempty"`
	Granted []Capability `json:"granted,omitempty"`
}

type InstallResponse struct {
	Component        string     `json:"component"`
	AlreadyInstalled bool       `json:"already_installed,omitempty"`
	Skipped          bool       `json:"skipped,omitempty"`
	Mutations        []Mutation `json:"mutations,omitempty"`
}

type CheckRequest struct{}

type CheckResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
	FixHint string `json:"fix_hint,omitempty"`
}

type UninstallRequest struct {
	DryRun         bool       `json:"dry_run,omitempty"`
	Global         bool       `json:"global,omitempty"`
	Project        bool       `json:"project,omitempty"`
	All            bool       `json:"all,omitempty"`
	PriorMutations []Mutation `json:"prior_mutations,omitempty"`
}

type UninstallResponse struct {
	Component string     `json:"component"`
	Skipped   bool       `json:"skipped,omitempty"`
	Mutations []Mutation `json:"mutations,omitempty"`
}

type HookRequest struct {
	HookName string            `json:"hook_name"`
	Payload  json.RawMessage   `json:"payload,omitempty"`
	Env      map[string]string `json:"env,omitempty"`
}

type HookResponse struct {
	OK      bool            `json:"ok"`
	Result  json.RawMessage `json:"result,omitempty"`
	Message string          `json:"message,omitempty"`
}

// Handler types each method exposes.
type (
	DetectHandler    func(context.Context, *DetectRequest) (*DetectResponse, error)
	InstallHandler   func(context.Context, *InstallRequest) (*InstallResponse, error)
	CheckHandler     func(context.Context, *CheckRequest) (*CheckResponse, error)
	UninstallHandler func(context.Context, *UninstallRequest) (*UninstallResponse, error)
	HookHandler      func(context.Context, *HookRequest) (*HookResponse, error)
)

// Server is the host-side bridge. Methods are unset by default; a
// nil method handler returns an "unimplemented" error to the caller.
type Server struct {
	socketPath string
	listener   net.Listener
	mu         sync.RWMutex
	detect     DetectHandler
	install    InstallHandler
	check      CheckHandler
	uninstall  UninstallHandler
	hooks      map[string]HookHandler
	wg         sync.WaitGroup
	stopped    bool
}

// NewServer constructs a server bound to socketPath. Call Serve to
// start accepting connections.
func NewServer(socketPath string) *Server {
	return &Server{socketPath: socketPath, hooks: map[string]HookHandler{}}
}

// HandleDetect registers the Detect handler.
func (s *Server) HandleDetect(h DetectHandler) { s.mu.Lock(); s.detect = h; s.mu.Unlock() }

// HandleInstall registers the Install handler.
func (s *Server) HandleInstall(h InstallHandler) { s.mu.Lock(); s.install = h; s.mu.Unlock() }

// HandleCheck registers the Check handler.
func (s *Server) HandleCheck(h CheckHandler) { s.mu.Lock(); s.check = h; s.mu.Unlock() }

// HandleUninstall registers the Uninstall handler.
func (s *Server) HandleUninstall(h UninstallHandler) { s.mu.Lock(); s.uninstall = h; s.mu.Unlock() }

// HandleHook registers a per-name hook handler.
func (s *Server) HandleHook(name string, h HookHandler) {
	s.mu.Lock()
	s.hooks[name] = h
	s.mu.Unlock()
}

// Listen binds the socket and returns its absolute path. Useful for
// tests that need to look up the socket before Serve runs.
func (s *Server) Listen() (string, error) {
	if err := os.MkdirAll(filepath.Dir(s.socketPath), 0o700); err != nil {
		return "", err
	}
	_ = os.Remove(s.socketPath)
	l, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return "", fmt.Errorf("bridge: listen: %w", err)
	}
	s.listener = l
	return s.socketPath, nil
}

// Serve accepts and dispatches connections until Stop is called.
func (s *Server) Serve() error {
	if s.listener == nil {
		if _, err := s.Listen(); err != nil {
			return err
		}
	}
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.mu.RLock()
			stopped := s.stopped
			s.mu.RUnlock()
			if stopped {
				return nil
			}
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}
		s.wg.Add(1)
		go func(c net.Conn) {
			defer s.wg.Done()
			s.handle(c)
		}(conn)
	}
}

// Stop closes the listener and waits for in-flight handlers.
func (s *Server) Stop() error {
	s.mu.Lock()
	s.stopped = true
	s.mu.Unlock()
	if s.listener != nil {
		_ = s.listener.Close()
	}
	s.wg.Wait()
	_ = os.Remove(s.socketPath)
	return nil
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	rd := bufio.NewReader(conn)
	body, err := io.ReadAll(rd)
	if err != nil {
		return
	}
	var req Request
	if err := json.Unmarshal(body, &req); err != nil {
		writeErr(conn, fmt.Errorf("bridge: malformed request: %w", err))
		return
	}
	resp, err := s.dispatch(context.Background(), &req)
	if err != nil {
		writeErr(conn, err)
		return
	}
	out, _ := json.Marshal(resp)
	_, _ = conn.Write(out)
}

func (s *Server) dispatch(ctx context.Context, req *Request) (*Response, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	switch req.Method {
	case MethodDetect:
		if s.detect == nil {
			return nil, fmt.Errorf("unimplemented: %s", req.Method)
		}
		var inner DetectRequest
		_ = json.Unmarshal(req.Body, &inner)
		out, err := s.detect(ctx, &inner)
		return wrap(out, err)
	case MethodInstall:
		if s.install == nil {
			return nil, fmt.Errorf("unimplemented: %s", req.Method)
		}
		var inner InstallRequest
		_ = json.Unmarshal(req.Body, &inner)
		out, err := s.install(ctx, &inner)
		return wrap(out, err)
	case MethodCheck:
		if s.check == nil {
			return nil, fmt.Errorf("unimplemented: %s", req.Method)
		}
		var inner CheckRequest
		_ = json.Unmarshal(req.Body, &inner)
		out, err := s.check(ctx, &inner)
		return wrap(out, err)
	case MethodUninstall:
		if s.uninstall == nil {
			return nil, fmt.Errorf("unimplemented: %s", req.Method)
		}
		var inner UninstallRequest
		_ = json.Unmarshal(req.Body, &inner)
		out, err := s.uninstall(ctx, &inner)
		return wrap(out, err)
	case MethodHook:
		var inner HookRequest
		_ = json.Unmarshal(req.Body, &inner)
		h, ok := s.hooks[inner.HookName]
		if !ok {
			return nil, fmt.Errorf("unknown hook: %s", inner.HookName)
		}
		out, err := h(ctx, &inner)
		return wrap(out, err)
	default:
		return nil, fmt.Errorf("unknown method: %s", req.Method)
	}
}

func wrap(payload any, err error) (*Response, error) {
	if err != nil {
		return nil, err
	}
	body, _ := json.Marshal(payload)
	return &Response{OK: true, Body: body}, nil
}

func writeErr(w io.Writer, err error) {
	body, _ := json.Marshal(Response{OK: false, Error: err.Error()})
	_, _ = w.Write(body)
}

// Client is the plugin-side counterpart. It speaks the same protocol
// over the same socket; OCI plugins import a generated client (or the
// bridge.Client when written in Go) and call .Detect / .Install / etc.
type Client struct {
	socketPath string
}

// NewClient constructs a client targeting socketPath.
func NewClient(socketPath string) *Client { return &Client{socketPath: socketPath} }

// Call invokes one method and decodes the response into out.
func (c *Client) Call(ctx context.Context, method string, in, out any) error {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return fmt.Errorf("bridge: dial %s: %w", c.socketPath, err)
	}
	defer conn.Close()
	body, _ := json.Marshal(in)
	req := Request{Method: method, Body: body}
	rb, _ := json.Marshal(req)
	if _, err := conn.Write(rb); err != nil {
		return err
	}
	if cw, ok := conn.(interface{ CloseWrite() error }); ok {
		_ = cw.CloseWrite()
	}
	respBody, err := io.ReadAll(conn)
	if err != nil {
		return err
	}
	var resp Response
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("bridge: %s: %s", method, resp.Error)
	}
	if out != nil {
		return json.Unmarshal(resp.Body, out)
	}
	return nil
}
