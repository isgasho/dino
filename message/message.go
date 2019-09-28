package message

import (
	"fmt"
	"math/rand"
	"unicode"
)

type Kind uint8

const (
	// KindGet is a message from the client to the server, stating the client wants
	// to know the latest version of the value for a given key. It is never sent
	// from the server to the client. This kind of message should only be issued
	// when the client does not have a version of the value for the given key, or
	// the client knows the version is stale, e.g., because of a put error.
	KindGet Kind = iota

	// KindPut is a message that can be sent both by the client and the server. It
	// is used by clients to update a key's corresponding value with a new version.
	// The server responds with the exact same put message if the put is accepted,
	// or with and error message. The server also fans out accepted put messages to
	// all clients that are connected, so they can keep up to date. This way,
	// clients should not need to issue get messages often.
	KindPut

	// KindError is a message that is only sent from the server to the client. That
	// can be in response to a get message (in case the requested key is not known)
	// or in response to a put message (in case the put version number is wrong).
	// The version number in a put message should match the one in the server,
	// proving that the client is up to date. If the put is stale, the client may
	// be overwriting some value, so the client should get the latest version and
	// possibly redo the put with the correct version, or give up the put). Other
	// error conditions might arise.
	KindError
)

// String implements fmt.Stringer.
func (k Kind) String() string {
	switch k {
	case KindGet:
		return "GET"
	case KindPut:
		return "PUT"
	case KindError:
		return "ERROR"
	default:
		return "unknown message kind"
	}
}

type Message struct {
	// The kind of the message. Meaningful for all messages.
	kind Kind

	// Correlates requests with responses for a given client. (Surely one won't
	// have more than 65336 requests waiting for a response?) Messages from other
	// clients will be tagged zero. Meaningful for all messages. The zero tag is
	// reserved for broadcast messages (those that are not responses to requests).
	tag uint16

	// The key to get or put. Meaningful for get and put messages only.
	key string

	// The value for a put message; doubles as a textual description of the error
	// for error messages.
	value string

	// Version of the value. Meaningful only for put messages.
	version uint64
}

func repr(any string) string {
	const max = 11
	for i, r := range any {
		if r > unicode.MaxASCII || !unicode.IsPrint(r) {
			// Not printable.
			return repr(fmt.Sprintf("%x", any))
		}
		if i > max {
			// Printable, but too long.
			return any[:max-3] + "..."
		}
	}
	// Printable and short!
	return any
}

// String implements fmt.Stringer. Note that the fmt package will try to call
// Error() first (and panic if it's not an error message). Should perhaps not
// make Message implement error.
func (m Message) String() string {
	return fmt.Sprintf("kind=%v tag=%d key=%s value=%s version=%d",
		m.kind, m.tag, repr(m.key), repr(m.value), m.version)
}

func (m Message) Kind() Kind {
	return m.kind
}

func (m Message) Tag() uint16 {
	return m.tag
}

func (m Message) Key() string {
	switch m.kind {
	case KindGet, KindPut:
		return m.key
	default:
		panic(m.accessorPanic("Key"))
	}
}

func (m Message) Value() string {
	switch m.kind {
	case KindError, KindPut:
		return m.value
	default:
		panic(m.accessorPanic("Value"))
	}
}

func (m Message) Version() uint64 {
	switch m.kind {
	case KindPut:
		return m.version
	default:
		panic(m.accessorPanic("Version"))
	}
}

func (m Message) accessorPanic(accessorName string) string {
	return fmt.Sprintf("cannot call .%s for message of kind %v", accessorName, m.kind)
}

func NewGetMessage(tag uint16, key string) Message {
	return Message{
		kind: KindGet,
		tag:  tag,
		key:  key,
	}
}

func NewPutMessage(tag uint16, key string, value string, version uint64) Message {
	return Message{
		kind:    KindPut,
		tag:     tag,
		key:     key,
		value:   value,
		version: version,
	}
}

func NewErrorMessage(tag uint16, message string) Message {
	return Message{
		kind:  KindError,
		tag:   tag,
		value: message,
	}
}

// ForBroadcast returns a copy of the message that's suitable to be broadcast to
// many connections.
func (m Message) ForBroadcast() Message {
	if m.kind != KindPut {
		panic(fmt.Sprintf("attempting to broadcast a message of kind: %v", m.kind))
	}
	m.tag = 0
	return m
}

func RandomTag() uint16 {
	return uint16(rand.Int() % 65536)
}

func RandomBytes() []byte {
	size := rand.Int() % 64
	key := make([]byte, size)
	rand.Read(key)
	return key
}

func RandomString() string {
	return string(RandomBytes())
}

func RandomVersion() uint64 {
	return rand.Uint64()
}
