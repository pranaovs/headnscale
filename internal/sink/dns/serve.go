package dns

import "github.com/miekg/dns"

func (s *Sink) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	s.mu.RLock()
	nodes := s.nodes
	s.mu.RUnlock()
}
