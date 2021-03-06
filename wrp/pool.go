package wrp

import (
	"io"
)

const (
	DefaultPoolSize = 100
)

// EncoderPool represents a pool of Encoder objects that can be used as is
// encode WRP messages.  Unlike a sync.Pool, this pool holds on to its pooled
// encoders across garbage collections.
type EncoderPool struct {
	pool   chan Encoder
	format Format
}

// NewEncoderPool returns an EncoderPool for a given format.  The initialBufferSize is
// used when encoding to byte arrays.  If this value is nonpositive, DefaultInitialBufferSize
// is used instead.
func NewEncoderPool(poolSize int, f Format) *EncoderPool {
	if poolSize < 1 {
		poolSize = DefaultPoolSize
	}

	ep := &EncoderPool{
		pool:   make(chan Encoder, poolSize),
		format: f,
	}

	for repeat := 0; repeat < poolSize; repeat++ {
		ep.pool <- ep.New()
	}

	return ep
}

// Format returns the wrp format this pool encodes to
func (ep *EncoderPool) Format() Format {
	return ep.format
}

// New simply creates a new Encoder using this pool's configuration.
// This method is used internally to populate and manage the pool, but
// can also be used externally to obtain a new, unpooled instance.
func (ep *EncoderPool) New() Encoder {
	return NewEncoder(nil, ep.format)
}

// Get returns an Encoder from the pool.  If the pool is empty, a new Encoder is
// created using the initial pool configuration.  This method never returns nil.
func (ep *EncoderPool) Get() (encoder Encoder) {
	select {
	case encoder = <-ep.pool:
	default:
		encoder = ep.New()
	}

	return
}

// Put returns an Encoder to the pool.  If this pool is full or if the supplied
// encoder is nil, this method does nothing.
func (ep *EncoderPool) Put(encoder Encoder) {
	if encoder != nil {
		select {
		case ep.pool <- encoder:
		default:
		}
	}
}

// Encode uses an Encoder from the pool to encode the source into the destination
func (ep *EncoderPool) Encode(destination io.Writer, source interface{}) error {
	encoder := ep.Get()
	defer ep.Put(encoder)

	encoder.Reset(destination)
	return encoder.Encode(source)
}

// EncodeBytes uses an encoder from the pool to encode the source into a byte array.
// The destination pointer will be replaced with a slice sized for the encoded data,
// using a zero-copy approach.  If destination has points to a slice with adequate capacity,
// no new memory allocation is done.
func (ep *EncoderPool) EncodeBytes(destination *[]byte, source interface{}) error {
	encoder := ep.Get()
	defer ep.Put(encoder)

	encoder.ResetBytes(destination)
	return encoder.Encode(source)
}

// DecoderPool is a pool of Decoder instances for a specific format
type DecoderPool struct {
	pool   chan Decoder
	format Format
}

// NewDecoderPool returns a DecoderPool that works with a given Format
func NewDecoderPool(poolSize int, f Format) *DecoderPool {
	if poolSize < 1 {
		poolSize = DefaultPoolSize
	}

	dp := &DecoderPool{
		pool:   make(chan Decoder, poolSize),
		format: f,
	}

	for repeat := 0; repeat < poolSize; repeat++ {
		dp.pool <- dp.New()
	}

	return dp
}

// Format returns the wrp format this pool decodes from
func (ep *DecoderPool) Format() Format {
	return ep.format
}

// New simply creates a new Decoder using this pool's configuration.
// This method is used internally to populate and manage the pool, but
// can also be used externally to obtain a new, unpooled instance.
func (ep *DecoderPool) New() Decoder {
	return NewDecoder(nil, ep.format)
}

// Get obtains a Decoder from the pool.  If the pool is empty, a new Decoder is
// created using the initial pool configuration.  This method never returns nil.
func (dp *DecoderPool) Get() (decoder Decoder) {
	select {
	case decoder = <-dp.pool:
	default:
		decoder = dp.New()
	}

	return
}

// Put returns a Decoder to the pool.  If this pool is full or if the supplied
// decoder is nil, this method does nothing.
func (dp *DecoderPool) Put(decoder Decoder) {
	if decoder != nil {
		select {
		case dp.pool <- decoder:
		default:
		}
	}
}

// Decode unmarshals data from the source onto the destination instance, which is
// normally a pointer to some struct (such as *Message).
func (dp *DecoderPool) Decode(destination interface{}, source io.Reader) error {
	decoder := dp.Get()
	defer dp.Put(decoder)

	decoder.Reset(source)
	return decoder.Decode(destination)
}

// DecodeBytes unmarshals data from the source byte slice onto the destination instance.
// The destination is typically a pointer to a struct, such as *Message.
func (dp *DecoderPool) DecodeBytes(destination interface{}, source []byte) error {
	decoder := dp.Get()
	defer dp.Put(decoder)

	decoder.ResetBytes(source)
	return decoder.Decode(destination)
}
