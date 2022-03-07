package network

type Packet struct {
	src     Endpoint
	dest    Endpoint
	payload interface{}
}

func NewPacket(src, dest Endpoint, payload interface{}) *Packet {
	return &Packet{
		src:     src,
		dest:    dest,
		payload: payload,
	}
}

func (p Packet) Src() Endpoint {
	return p.src
}

func (p Packet) Dest() Endpoint {
	return p.dest
}

func (p Packet) Payload() interface{} {
	return p.payload
}
