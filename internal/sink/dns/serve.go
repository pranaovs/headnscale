package dns

import (
	"log"
	"strings"

	"github.com/miekg/dns"
	"github.com/pranaovs/headnscale/internal/types"
)

func (s *Sink) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true

	if len(r.Question) == 0 {
		_ = w.WriteMsg(msg)
		return
	}

	s.mu.RLock()
	nodes := s.nodes
	s.mu.RUnlock()

	for _, question := range r.Question {
		ip, found := questionToNodeIP(question.Name, s.BaseDomain, s.NoBaseDomain, nodes)
		if !found {
			continue
		}

		// At this point, we definitely have a node that matches the question
		switch question.Qtype {
		case dns.TypeA:
			if ip.IPv4 == nil {
				continue
			}

			response := new(dns.A)
			response.Hdr = dns.RR_Header{
				Name:   question.Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    60,
			}
			response.A = ip.IPv4
			msg.Answer = append(msg.Answer, response)

		case dns.TypeAAAA:
			if ip.IPv6 == nil {
				continue
			}

			response := new(dns.AAAA)
			response.Hdr = dns.RR_Header{
				Name:   question.Name,
				Rrtype: dns.TypeAAAA,
				Class:  dns.ClassINET,
				Ttl:    60,
			}
			response.AAAA = ip.IPv6
			msg.Answer = append(msg.Answer, response)
		}
	}

	if err := w.WriteMsg(msg); err != nil {
		log.Printf("DNS response write error: %v", err)
	}
}

func questionToNodeIP(questionName string, BaseDomain string, noBaseDomain bool, nodes []types.Node) (types.IP, bool) {
	questionName = strings.ToLower(strings.TrimSuffix(questionName, "."))
	BaseDomain = strings.ToLower(strings.TrimSuffix(BaseDomain, "."))

	// First pass, exact match
	for _, node := range nodes {
		nodeName := strings.ToLower(node.Hostname)
		if noBaseDomain && nodeName == questionName {
			return node.IP, true
		}
		if BaseDomain != "" {
			fqdn := nodeName + "." + BaseDomain
			if fqdn == questionName {
				return node.IP, true
			}
		}
	}

	// Second pass, wildcard check
	for _, node := range nodes {
		nodeName := strings.ToLower(node.Hostname)
		if BaseDomain != "" {
			fqdn := nodeName + "." + BaseDomain
			if dns.IsSubDomain(fqdn+".", questionName+".") && questionName != fqdn {
				return node.IP, true
			}
		}
		if noBaseDomain {
			if dns.IsSubDomain(nodeName+".", questionName+".") && questionName != nodeName {
				return node.IP, true
			}
		}
	}
	return types.IP{}, false
}
