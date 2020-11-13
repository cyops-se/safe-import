package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
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
	ip, err := externalIP()
	if err != nil {
		fmt.Println(err)
	}

	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			if ip != "" {
				var rr, empty dns.RR
				if rr, err = dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip)); err != nil {
					fmt.Println("DNS CRITICAL: failed to create A RR, err:", err)
					return
				}

				if empty, err = dns.NewRR(fmt.Sprintf("%s IN A", q.Name)); err != nil {
					fmt.Println("DNS CRITICAL: failed to create EMTPY A RR, err:", err)
					return
				}

				// Send the URI over nats if used
				ip, _ := w.RemoteAddr().(*net.UDPAddr)
				data := &types.DnsRequest{Time: time.Now().UTC(), FromIP: ip.IP.String(), Query: q.Name,
					MatchQuery: regexp.QuoteMeta(q.Name), LastSeen: time.Now().UTC()}

				if match := svc.matchQuery(q.Name, "white"); match != nil {
					if match.Allowed {
						fmt.Println("WHITE ALLOWED!", data.Query, data.FromIP, match.Query, match.MatchQuery)
						m.Answer = append(m.Answer, rr)
					} else {
						fmt.Println("WHITE NOT ALLOWED!", data.Query, data.FromIP, match.Query, match.MatchQuery)
						m.Answer = append(m.Answer, empty)
						m.Rcode = dns.RcodeNameError
					}
				} else if match := svc.matchQuery(q.Name, "black"); match != nil {
					fmt.Println("BLACK ALERT! Bad DNS query detected", q.Name)
					m.Answer = append(m.Answer, empty)
					m.Rcode = dns.RcodeNameError
				} else if match := svc.checkGrey(q.Name); match == nil {
					m.Answer = append(m.Answer, rr)
					data.Class = "grey"
					data.Count = 1
					common.DB.Create(data)
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
func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}

		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
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
