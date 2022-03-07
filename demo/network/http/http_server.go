package http

import (
	"demo/network"
	"errors"
)

// Handler HTTP请求处理接口
type Handler func(req *Request) *Response

// Server Http服务器
type Server struct {
	socket        network.Socket
	localEndpoint network.Endpoint
	routers       map[Method]map[Uri]Handler
}

func NewServer(socket network.Socket) *Server {
	server := &Server{
		socket:  socket,
		routers: make(map[Method]map[Uri]Handler),
	}
	server.socket.AddListener(server)
	return server
}

func (s *Server) Listen(ip string, port int) *Server {
	s.localEndpoint = network.EndpointOf(ip, port)
	return s
}

func (s *Server) Start() error {
	return s.socket.Listen(s.localEndpoint)
}

func (s *Server) Shutdown() {
	s.socket.Close(s.localEndpoint)
}

func (s *Server) Get(uri Uri, handler Handler) *Server {
	if _, ok := s.routers[GET]; !ok {
		s.routers[GET] = make(map[Uri]Handler)
	}
	s.routers[GET][uri] = handler
	return s
}

func (s *Server) Post(uri Uri, handler Handler) *Server {
	if _, ok := s.routers[POST]; !ok {
		s.routers[POST] = make(map[Uri]Handler)
	}
	s.routers[POST][uri] = handler
	return s
}

func (s *Server) Put(uri Uri, handler Handler) *Server {
	if _, ok := s.routers[PUT]; !ok {
		s.routers[PUT] = make(map[Uri]Handler)
	}
	s.routers[PUT][uri] = handler
	return s
}

func (s *Server) Delete(uri Uri, handler Handler) *Server {
	if _, ok := s.routers[DELETE]; !ok {
		s.routers[DELETE] = make(map[Uri]Handler)
	}
	s.routers[DELETE][uri] = handler
	return s
}

func (s *Server) Handle(packet *network.Packet) error {
	req, ok := packet.Payload().(*Request)
	if !ok {
		return errors.New("invalid packet, not http request")
	}
	if req.IsInValid() {
		resp := ResponseOfId(req.ReqId()).
			AddStatusCode(StatusBadRequest).
			AddProblemDetails("uri or method is invalid")
		return s.socket.Send(network.NewPacket(packet.Dest(), packet.Src(), resp))
	}

	router, ok := s.routers[req.Method()]
	if !ok {
		resp := ResponseOfId(req.ReqId()).
			AddStatusCode(StatusMethodNotAllow).
			AddProblemDetails(StatusMethodNotAllow.Details)
		return s.socket.Send(network.NewPacket(packet.Dest(), packet.Src(), resp))
	}

	var handler Handler
	for u, h := range router {
		if req.Uri().Contains(u) {
			handler = h
			break
		}
	}
	if handler == nil {
		resp := ResponseOfId(req.ReqId()).
			AddStatusCode(StatusNotFound).
			AddProblemDetails("can not find handler of uri")
		return s.socket.Send(network.NewPacket(packet.Dest(), packet.Src(), resp))
	}

	resp := handler(req)
	return s.socket.Send(network.NewPacket(packet.Dest(), packet.Src(), resp))
}
