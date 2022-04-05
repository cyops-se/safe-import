package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cyops-se/safe-import/si-inner/common"
	"github.com/cyops-se/safe-import/si-inner/types"
	"github.com/cyops-se/safe-import/usvc"
	"github.com/miekg/dns"
)

type DnsQueryMessage struct {
	IPAddress string `json:"ip"`
	Query     string `json:"query"`
}

type InnerDnsService struct {
	usvc.Usvc
}

func (svc *InnerDnsService) Initialize(broker *usvc.UsvcBroker) {
	svc.InitializeService(broker, 1, "si-inner", "dns", "An inner part handling DNS requests for the safe-import solution")
	svc.RegisterMethod("allitems", svc.allItems)
	svc.RegisterMethod("byfieldname", svc.byFieldName)
	svc.RegisterMethod("update", svc.update)
	svc.RegisterMethod("prune", svc.prune)
	svc.RegisterMethod("delete", svc.delete)

	// We don't use settings right now
	if err := svc.LoadSettings(); err != nil {
		svc.SaveSettings() // Save default settings. Though we don't actually use the settings right now...
	}

	svc.Executor = svc.execute
	svc.SetTaskIdleTime(60 * 1) // every minute
	svc.execute()

	go svc.DNSServer()
}

// This code was taken from the README on the package "github.com/miekg/dns"
// but modified to handle all requests
func (svc *InnerDnsService) DNSServer() {
	// attach request handler func
	dns.HandleFunc(".", svc.handleDnsRequest)

	// start server
	port := 53
	server := &dns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}
	svc.LogGeneric("info", "Starting DNS server at port :%d", port)
	err := server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		svc.LogGeneric("info", "Failed to start DNS server: %#v", err.Error())
	}
}

func (svc *InnerDnsService) parseQuery(m *dns.Msg, w dns.ResponseWriter) {
	remoteaddr, _ := w.RemoteAddr().(*net.UDPAddr)
	remoteip := remoteaddr.IP

	dnsip, err := externalIP(remoteip)
	if err != nil {
		svc.LogError("Failed to find external IP (to use as DNS response)", err)
		return
	}

	// svc.LogInfo(fmt.Sprintf("Using %s as IP address for DNS response", dnsip))

	for _, q := range m.Question {
		switch q.Qtype {
		default:
			m.Rcode = dns.RcodeNameError
		case dns.TypeAAAA, dns.TypeA:
			if dnsip != "" {
				var rr dns.RR
				if q.Qtype == dns.TypeA {
					if rr, err = dns.NewRR(fmt.Sprintf("%s A %s", q.Name, dnsip)); err != nil {
						m.Rcode = dns.RcodeNameError
						return
					}
				} else {
					if rr, err = dns.NewRR(fmt.Sprintf("%s AAAA ::ffff:%s", q.Name, dnsip)); err != nil {
						m.Rcode = dns.RcodeNameError
						return
					}
				}

				// Send the URI over nats if used
				// ip, _ := w.RemoteAddr().(*net.UDPAddr)
				data := &types.DnsRequest{Time: time.Now().UTC(), FromIP: remoteip.String(), Query: q.Name,
					MatchQuery: regexp.QuoteMeta(q.Name), LastSeen: time.Now().UTC()}

				if match := svc.matchQuery(q.Name, "white"); match != nil {
					if match.Allowed {
						svc.LogGeneric("info", "inner: WHITE DNS ALLOWED: %s, from %s", data.Query, data.FromIP)
						m.Answer = append(m.Answer, rr)
					} else {
						svc.LogGeneric("info", "inner: WHITE DNS BLOCKED: %s, from %s", data.Query, data.FromIP)
						m.Rcode = dns.RcodeNameError
					}
				} else if match := svc.matchQuery(q.Name, "black"); match != nil {
					svc.LogGeneric("alert", "BLACK DNS ALERT: %s, from %s", data.Query, data.FromIP)
					m.Rcode = dns.RcodeNameError
				} else {
					svc.LogGeneric("info", "inner: GREY DNS BLOCKED: %s, from %s", data.Query, data.FromIP)
					m.Rcode = dns.RcodeNameError
					if match := svc.checkGrey(q.Name); match == nil {
						data.Class = "grey"
						data.Count = 1
						common.DB.Create(data)
					}
				}
			}
		case dns.TypeSRV: // Only supports HTTPS at the moment
			if dnsip != "" {
				var addrs []*net.SRV
				name := strings.Replace(q.Name, "_https._tcp.", "", 1)
				_, addrs, err = net.LookupSRV("https", "tcp", name)
				if err != nil {
					// svc.LogGeneric("error", "DNS ERROR: failed to lookup SRV record %s, err: %s", q.Name, err.Error())
					m.Rcode = dns.RcodeNameError
					return
				}

				var rr dns.RR
				if rr, err = dns.NewRR(fmt.Sprintf("%s SRV %d %d %d %s", q.Name, addrs[0].Priority, addrs[0].Weight, addrs[0].Port, addrs[0].Target)); err != nil {
					m.Rcode = dns.RcodeNameError
					return
				}

				// Send the URI over nats if used
				data := &types.DnsRequest{Time: time.Now().UTC(), FromIP: remoteip.String(), Query: q.Name,
					MatchQuery: regexp.QuoteMeta(q.Name), LastSeen: time.Now().UTC()}

				if match := svc.matchQuery(q.Name, "white"); match != nil {
					if match.Allowed {
						svc.LogGeneric("info", "inner: WHITE DNS ALLOWED: %s, from %s", data.Query, data.FromIP)
						m.Answer = append(m.Answer, rr)
					} else {
						svc.LogGeneric("info", "inner: WHITE DNS BLOCKED: %s, from %s", data.Query, data.FromIP)
						m.Rcode = dns.RcodeNameError
					}
				} else if match := svc.matchQuery(q.Name, "black"); match != nil {
					svc.LogGeneric("alert", "inner: BLACK DNS ALERT: %s, from %s", data.Query, data.FromIP)
					m.Rcode = dns.RcodeNameError
				} else {
					svc.LogGeneric("info", "inner: GREY DNS BLOCKED: %s, from %s", data.Query, data.FromIP)
					m.Rcode = dns.RcodeNameError
					if match := svc.checkGrey(q.Name); match == nil {
						data.Class = "grey"
						data.Count = 1
						common.DB.Create(data)
					}
				}
			}
		}
	}
}

func (svc *InnerDnsService) matchQuery(q string, class string) *types.DnsRequest {
	var existing []types.DnsRequest
	if result := common.DB.Find(&existing, "class LIKE ?", class); result.RowsAffected <= 0 {
		return nil
	}

	// Now we iterate through the entries that matches scheme and fqdn
	// and try to find a match using the entry MatchURL regexp
	for _, v := range existing {
		re, _ := regexp.Compile(v.MatchQuery)
		if re.MatchString(q) {
			v.Count++
			v.LastSeen = time.Now().UTC()
			common.DB.Save(v)
			return &v
		}
	}

	return nil
}

func (svc *InnerDnsService) checkGrey(q string) *types.DnsRequest {
	var existing types.DnsRequest
	if result := common.DB.First(&existing, "query = ? AND class = 'grey'", q); result.RowsAffected <= 0 {
		return nil
	}

	existing.Count++
	existing.LastSeen = time.Now().UTC()
	common.DB.Save(&existing)

	return &existing
}

func (svc *InnerDnsService) handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		svc.parseQuery(m, w)
	}

	w.WriteMsg(m)
}

// Taken from "https://play.golang.org/p/BDt3qEQ_2H"
func externalIP(remoteip net.IP) (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Printf("externalIP:net.Interfaces() failed, error: %s", err.Error())
		return "", err
	}

	for _, iface := range ifaces {
		// // fmt.Println("Checking interface", iface)
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}

		if (iface.Flags & net.FlagLoopback) != 0 {
			continue // loopback interface
		}

		addrs, err := iface.Addrs()
		if err != nil {
			log.Printf("externalIP:iface.Addrs() failed, error: %s", err.Error())
			return "", err
		}

		for _, addr := range addrs {
			var ip net.IP
			// var mask net.IPMask

			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
				// mask = v.Mask
				// // fmt.Println("NETWORK", ip, mask)
			case *net.IPAddr:
				ip = v.IP
				// mask = v.IP.DefaultMask()
				// // fmt.Println("IP", ip, mask)
			}

			ip = ip.To4()
			if ip == nil || ip.IsLoopback() {
				continue // not an ipv4 address or loopback
			}

			return ip.String(), nil
		}
	}

	return "", errors.New("are you connected to the network?")
}

func (svc *InnerDnsService) execute() {
}

func (svc *InnerDnsService) allItems(payload string) (interface{}, error) {
	var items []types.DnsRequest
	common.DB.Find(&items)
	return items, nil
}

func (svc *InnerDnsService) byFieldName(payload string) (interface{}, error) {
	var args types.ByNameRequest
	if err := json.Unmarshal([]byte(payload), &args); err != nil {
		svc.LogGeneric("error", "Marshalling request to JSON failed: %#v", err)
		return nil, err
	}

	var items []types.DnsRequest
	if result := common.DB.Where(map[string]interface{}{args.Name: args.Value}).Find(&items); result.Error != nil {
		svc.LogGeneric("error", "Failed to query database, error: %#v", result.Error)
		return nil, result.Error
	}

	return items, nil
}

func (svc *InnerDnsService) update(payload string) (interface{}, error) {
	var item types.DnsRequest
	if err := json.Unmarshal([]byte(payload), &item); err != nil {
		svc.LogGeneric("error", "Marshalling request to JSON failed: %#v", err)
		return nil, err
	}

	if result := common.DB.Save(&item); result.Error != nil {
		svc.LogGeneric("error", "Failed to query database, error: %#v", result.Error)
		return nil, result.Error
	}

	return item, nil
}

func (svc *InnerDnsService) prune(payload string) (interface{}, error) {
	var greyitems, whiteitems, blackitems []types.DnsRequest

	// Get all lists
	if result := common.DB.Find(&greyitems, "class = 'grey'"); result.Error != nil {
		svc.LogGeneric("error", "Failed to query database, error: %#v", result.Error)
		return nil, result.Error
	}
	if result := common.DB.Find(&whiteitems, "class = 'white'"); result.Error != nil {
		svc.LogGeneric("error", "Failed to query database, error: %#v", result.Error)
		return nil, result.Error
	}
	if result := common.DB.Find(&blackitems, "class = 'black'"); result.Error != nil {
		svc.LogGeneric("error", "Failed to query database, error: %#v", result.Error)
		return nil, result.Error
	}

	// Pruning entries means removing all entries from the grey list that
	// matches a MatchURL regexp field of a white or black entry.
	for _, wv := range whiteitems {
		for _, gv := range greyitems {
			re, _ := regexp.Compile(wv.MatchQuery)
			if re.MatchString(gv.Query) {
				if result := common.DB.Delete(&gv); result.Error != nil {
					svc.LogGeneric("error", "Failed to delete item from database, error: %#v", result.Error)
					return nil, result.Error
				}
			}
		}
	}

	for _, bv := range blackitems {
		for _, gv := range greyitems {
			re, _ := regexp.Compile(bv.MatchQuery)
			if re.MatchString(gv.Query) {
				if result := common.DB.Delete(&gv); result.Error != nil {
					svc.LogGeneric("error", "Failed to delete item from database, error: %#v", result.Error)
					return nil, result.Error
				}
			}
		}
	}

	return nil, nil
}

func (svc *InnerDnsService) delete(payload string) (interface{}, error) {
	var id string
	if err := json.Unmarshal([]byte(payload), &id); err != nil {
		svc.LogGeneric("error", "Marshalling request to JSON failed: %#v", err)
		return nil, err
	}

	if result := common.DB.Delete(&types.DnsRequest{}, id); result.Error != nil {
		svc.LogGeneric("error", "Failed to query database, error: %#v", result.Error)
		return nil, result.Error
	}

	return nil, nil
}
