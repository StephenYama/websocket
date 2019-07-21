// Package websocket implements the WebSocket protocol defined in RFC 6455.
//
// Overview
//
// The Conn type represents a WebSocket connection. Use Dial to create a
// connection on a client and Upgrade to create a connection on a server.
//
// Here's an example showing how to use Upgrade on the server:
//
//  var upgradeOptions = websocket.UpgradeOptions{}
//
//  func handler(w http.ResponseWriter, r *http.Request) {
//      conn, err := websocket.Upgrade(w, r, nil, &upgradeOptions)
//      if err != nil {
//         log.Println(err)
//        return
//      }
//      XXX
//  }
//
package websocket

// CloseCode is a WebSocket close code.
type CloseCode int

//go:generate go run golang.org/x/tools/cmd/stringer -type=StatusCode

// Close codes defined in RFC 6455, section 11.7.
const (
	CloseNormalClosure           CloseCode = 1000
	CloseGoingAway               CloseCode = 1001
	CloseProtocolError           CloseCode = 1002
	CloseUnsupportedData         CloseCode = 1003
	CloseNoStatusReceived        CloseCode = 1005
	CloseAbnormalClosure         CloseCode = 1006
	CloseInvalidFramePayloadData CloseCode = 1007
	ClosePolicyViolation         CloseCode = 1008
	CloseMessageTooBig           CloseCode = 1009
	CloseMandatoryExtension      CloseCode = 1010
	CloseInternalServerErr       CloseCode = 1011
	CloseServiceRestart          CloseCode = 1012
	CloseTryAgainLater           CloseCode = 1013
	CloseTLSHandshake            CloseCode = 1015
)

// CloseError represents a close message received from a peer.
type CloseError struct {
	// Code is defined in RFC 6455, section 11.7.
	Code int

	// Reason is the optional text payload.
	Reason string
}

func (e *CloseError) Error() string {
	return fmt.Printf("websocket close: code = %v (%d), message = %s", e.Code, e.Code, e.Text)
}

// UpgradeOptions specifies options for upgrading an HTTP connection to a
// WebSocket connection.
type UpgradeOptions struct {
	// Error specifies the function for generating HTTP error responses. If
	// Error is nil, then http.Error is used to generate the HTTP response.
	Error func(w http.ResponseWriter, r *http.Request, status int, reason error)

	//
	// CheckOrigin returns true if the request Origin header is acceptable. If
	// CheckOrigin is nil, then a safe default is used: return false if the
	// Origin request header is present and the origin host is not equal to
	// request Host header.
	//
	// A CheckOrigin function should carefully validate the request origin to
	// prevent cross-site request forgery.
	OriginTestHandledByApplication bool

	// Subprotocols specifies the server's supported protocols in order of
	// preference. If this field is not nil, then the Upgrade method negotiates
	// a subprotocol by selecting the first match in this list with a protocol
	// requested by the client. If there's no match, then no protocol is
	// negotiated (the Sec-Websocket-Protocol header is not included in the
	// handshake response).
	Subprotocols []string

	// HandshakeTimeout specifies the duration for the handshake to complete.
	HandshakeTimeout time.Duration

	// The connection pings the peer when no data is received within
	// PingPeriod. A reasonable default is used when PingPeriod is zero.
	// Disable pinging by setting PingPeriod to a negative value.
	PingPeriod time.Duration

	// ReadLimit specifies the maximum size in bytes for a message read from
	// the peer. If a message exceeds the limit, the connection sends a close
	// message to the peer and returns ErrReadLimit to the application. If
	// ReadLimit is zero, then a limit of 32K is used. Set ReadLimit to a
	// negative number to disable limiting.  A call to Reader.SetLimit
	// overrides this limit.
	ReadLimit int64

	// ReadTimeout specifies a timeout for reading a message, starting at the
	// time that Conn.Reader returns a reader. A call to Reader.SetDeadline
	// overrides this timeout.
	ReadTimeout time.Duration

	// WriteTimeout specifies a timeout for writing a message, starting at the
	// time that Conn.Writer returns a writer. A call to Reader.SetDeadline
	// overrides this timeout.
	WriteTimeout time.Duration
}

// Upgrade upgrades the HTTP server connection to the WebSocket protocol.
//
// The responseHeader is included in the response to the client's upgrade
// request. Use the responseHeader to specify cookies (Set-Cookie) and the
// application negotiated subprotocol (Sec-WebSocket-Protocol).
//
// If the upgrade fails, then Upgrade replies to the client with an HTTP error
// response.
func Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header, options *UpgradeOptions) (*Conn, error)

// The Conn type represents a WebSocket connection.
type Conn struct {
}

// Subprotocol returns the negotiated protocol for the connection.
func (c *Conn) Subprotocol() string {}

// StartClose initiates the WebSocket closing handshake and arranges for
// Reader to timeout if a timely reply is not received. The application must
// call Reader to complete the closing handshake and close the connection.
//
// See the package documentation for more information on the closing handshake.
func (c *Conn) CloseWrite(ctx context.Context, code CloseCode, message string) error {}

// SetParentContext sets a parent context for the connection. The connection is
// closed when the parent context is canceled.
func (c *Conn) SetParentContextContext(ctx context.Context) context.Context {}

// Reader returns a Reader on the next data message received from the peer.
//
// Control messages are handled internerally by connection. The application must
// call Reaader in a loop to process these messages.
//
// The application must read each Reader until io.EOF or some other error is
// returned.
func (c *Conn) Reader(ctx context.Context) (Reader, error) { return Reader{}, nil }

// Writer returns a message writer. The application must close the writer when
// done writing the message.
func (c *Conn) Writer(ctx context.Context) (Writer, error) { return Writer{}, nil }

// Writer writes a message to the peer. Writer satisfies the io.Writer
// interface.
//
// A WebSocket message is written to the network as one or more frames. A frame
// has a header and application data. A final flag is set in the last frame
// header to indicate the last frame in a message.
//
// Each call to Write writes a frame to the underlying network connection. To
// reduce the overhead from frame headers and calls to the operating system,
// applications should avoid making many calls to Write with a small buffer.
//
// The SetFinal method sets the final flag on the next frame written by the
// Writer. Use this method to avoid sending an empty frame at the end of the
// message. The following code uses SetFinal to send a []byte as a single frame:
//
//  w, err := c.Writer()
//  if err != nil {
//      // handler error
//  }
//  w.SetFinal()
//  w.Write(data)
//  err := w.Close()
//  if err != nil {
//      // handle error
//  }
//
// Use the following code to minimize the number of frames written with a
// bufio.Writer:
//
//  w, err := c.Writer()
//  if err != nil {
//      // handler error
//  }
//  bw := bufio.NewWriter(w)
//  bw.Write(data)
//  ...
//  w.SetFinal()
//  bw.Flush()
//  err := w.Close()
//  if err != nil {
//      // handle error
//  }
type Writer struct {
	c     *Conn
	nonce int64
}

// Close ensures that final message frame is written to the network and
// releases resources used by the Writer.
//
// The application must close each writer.
func (w Writer) Close() error { return nil }

// SetBinary marks the message as a binary data message. Otherwise, the message
// is assumed to be a valid UTF-8 encoded text. SetBinary must be called before
// the first call to Write.
func (w Writer) SetBinary(binary bool) {}

// SetCompress determines whether the message is compressed when compression is
// negotiated with the peer. SetCompress must be called before the first call
// to Write. The default is to compress messages.
func (w Writer) SetCompress(compress bool) {}

// SetDeadline sets the deadline for future Write calls. A zero value for t
// means Write will not time out. SetDeadline overrides the timeout specified
// in DialOptions.MessageWriteTimeout and UpgradeOptions.MessageWriteTimeout.
func (w Writer) SetDeadline(t time.Time) error { return nil }

// Write writes p to the message. It returns the number of bytes written from p
// (0 <= n <= len(p)) and any error encountered that caused the write to stop
// early.
func (w Writer) Write(p []byte) (int, error) {}

// WriteString writes s to the message. It returns the number of bytes written from s
// (0 <= n <= len(s)) and any error encountered that caused the write to stop
// early.
func (w Writer) WriteString(s string) (int, error) {}

// SetFinal sets the final flag on the frame written by the next call to Write.
// This method optmizes the data written to the network. Applications do not
// need to call this method.
func (w Writer) SetFinal() {}

// Reader reads a message from the peer. MessageReader satisifies the io.Reader
// interface.
type Reader struct {
	c     *Conn
	nonce int64
}

// Binary returns true if the message is a WebSocket binary message.
// Otherwise, the message is a WebSocket TextMessage.
func (r Reader) Binary() bool {}

// Read reads up to len(p) bytes into p. It returns the number of bytes read (0
// <= n <= len(p)) and any error encountered.
func (r Reader) Read(p []byte) (int, error) {}

// SetDeadline sets the deadline for future Read calls. SetDeadline overrides
// the timeout set by ReadOptions.MessageTimeout  A zero value for t means Read
// will not time out.SetDeadline overrides the timeout specified in
// DialOptions.ReadTimeout and UpgradeOptions.ReadTimeout.
func (r Reader) SetDeadline(t time.Time) error {}

// SetLimit sets a limit on the number of bytes read in subsequent calls to
// Read. This limit overrides the limit specified in the RunOptions.ReadLimit
// field. When the limit is breached, a close message is sent to the
// peer and the connection is closed. A value of zero specifies no limit.
func (r Reader) SetLimit(n int64) {}

// ReadBytes reads the next message and returns it as a slice of bytes.
func ReadBytes(c *Conn, options *ReadOptions) (data []byte, isBinary bool, err error) {}

// ReadString reads the next message and returns it as a string.
func ReadString(c *Conn, options *ReadOptions) (data string, isBinary bool, err error) {}

// WriteBytes writes a slice of bytes as a WebSocket message.
func WriteBytes(c *Conn, options *WriteOptions, data []byte) error {}

// WriteString writes a string as a WebSocket message.
func WriteString(c *Conn, options *WriteOptions, data string) error {}

// WriteJSON encodes v as JSON and writes it as a message.
func WriteJSON(c *Conn, options *WriteOptions, v interface{}) error {}

// ReadJSON decodes the next received message as JSON to the value pointed to
// by v.
func ReadJSON(c *Conn, opitions *ReadOoptions, v interface{}) error {}
