package services

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	gtypes "github.com/cyops-se/safe-import/si-gatekeeper/types"
	"github.com/cyops-se/safe-import/si-inner/common"
	"github.com/cyops-se/safe-import/si-inner/types"
	"github.com/cyops-se/safe-import/usvc"
	"github.com/gorilla/mux"
)

var proxySvc *usvc.UsvcStub

type InnerHttpService struct {
	usvc.Usvc
}

func (svc *InnerHttpService) Initialize(broker *usvc.UsvcBroker) {
	svc.InitializeService(broker, 1, "si-inner", "http", "An inner part handling HTTP(S) requests for the safe-import solution")
	svc.RegisterMethod("allitems", svc.allItems)
	svc.RegisterMethod("byfieldname", svc.byFieldName)
	svc.RegisterMethod("update", svc.update)
	svc.RegisterMethod("prune", svc.prune)

	proxySvc = usvc.CreateStub(broker, "proxy", "si-gatekeeper", 1)

	// We don't use settings right now
	if err := svc.LoadSettings(); err != nil {
		svc.SaveSettings() // Save default settings. Though we don't actually use the settings right now...
	}

	svc.Executor = svc.execute
	svc.SetTaskIdleTime(60 * 1) // every minute
	svc.execute()

	go svc.startHTTPSServer()
	go svc.startHTTPServer()
}

// Starts the HTTPS server
func (svc *InnerHttpService) startHTTPSServer() {
	// Create a router for HTTPS
	r2 := mux.NewRouter()
	r2.PathPrefix("/").HandlerFunc(svc.handleHTTPSrequest)

	// Create HTTPS server with custom config
	s := &http.Server{
		Addr:      ":443",
		Handler:   r2,
		TLSConfig: new(tls.Config),
	}
	s.TLSConfig.GetCertificate = common.GetCertificateFunc

	svc.LogGeneric("info", "Starting HTTPS server at port :443")
	err := s.ListenAndServeTLS("", "")
	if err != nil {
		svc.LogGeneric("error", "Failed to start HTTPS server: %s", err.Error())
	}
}

// Starts the HTTP server
func (svc *InnerHttpService) startHTTPServer() {
	// Create a router for HTTP
	r := mux.NewRouter()
	r.PathPrefix("/").HandlerFunc(svc.handleHTTPrequest)

	svc.LogGeneric("info", "Starting HTTP server at port :80")
	err := http.ListenAndServe(":80", r)
	if err != nil {
		svc.LogGeneric("error", "Failed to start HTTP server: %s", err.Error())
	}
}

// Handles HTTPS requests pretty much identical to the one for HTTP
func (svc *InnerHttpService) handleHTTPSrequest(w http.ResponseWriter, r *http.Request) {
	svc.processHttp(w, r, "HTTPS")
}

// Handles HTTP requests
func (svc *InnerHttpService) handleHTTPrequest(w http.ResponseWriter, r *http.Request) {
	svc.processHttp(w, r, "HTTP")
}

func (svc *InnerHttpService) processHttp(w http.ResponseWriter, r *http.Request, scheme string) {
	// Pack the request in a custom structure for JSOn marshalling
	urlstring := fmt.Sprintf("%s://%s%s", strings.ToLower(scheme), r.Host, r.URL.RequestURI())
	msg := &types.HttpRequest{Time: time.Now().UTC(), Type: scheme, FromIP: strings.Split(r.RemoteAddr, ":")[0], ToHost: r.Host,
		Method: r.Method, URL: urlstring, MatchURL: regexp.QuoteMeta(urlstring), LastSeen: time.Now().UTC()}

	u, _ := url.Parse(msg.URL)
	// fmt.Println("Processing URL:", msg.URL)

	// See if the URL matches one in the lists (white, black, grey)
	if match := svc.matchURL(u, "white"); match != nil {
		if match.Allowed {
			fmt.Println("WHITE URL ALLOWED!", msg.URL, msg.FromIP)
			request := &gtypes.HttpDownloadRequest{msg.URL}
			if response, err := proxySvc.RequestMessage("httpget", request); err == nil {
				fmt.Println("RESPONSE request", response)
				if len(response.Payload) <= 0 {
					svc.LogError("Failed to download file via proxy", fmt.Errorf("Response payload from proxy download request is empty"))
				}

				var dr gtypes.HttpDownloadResponse
				if err := json.Unmarshal([]byte(response.Payload), &dr); err != nil {
					svc.LogGeneric("error", "Marshalling proxy response to JSON failed: %#v", err)
				}
				fmt.Println("RESPONSE payload", dr)
				if file, err := os.Open(dr.Filename); err == nil {
					defer file.Close()
					io.Copy(w, file)
				} else {
					svc.LogError("Failed to open file from si-outer", err)
				}
			} else {
				svc.LogError("Failed to request job from proxy", err)
			}
		} else {
			fmt.Println("WHITE URL NOT ALLOWED!", msg.URL, msg.FromIP)
		}
	} else if match := svc.matchURL(u, "black"); match != nil {
		fmt.Println("BLACK URL ALERT!", msg.URL, msg.FromIP)
	} else if match := svc.checkGrey(u); match == nil {
		msg.Class = "grey"
		msg.Count = 1
		common.DB.Create(msg)
	}
}

func (svc *InnerHttpService) matchURL(u *url.URL, class string) *types.HttpRequest {
	var existing []types.HttpRequest
	if result := common.DB.Find(&existing, "class LIKE ?", class); result.RowsAffected <= 0 {
		return nil
	}

	// Now we iterate through the entries that matches scheme and fqdn
	// and try to find a match using the entry MatchURL regexp
	for _, v := range existing {
		re, _ := regexp.Compile(v.MatchURL)
		if re.MatchString(u.String()) {
			v.Count++
			v.LastSeen = time.Now().UTC()
			common.DB.Save(v)
			return &v
		}
	}

	return nil
}

func (svc *InnerHttpService) checkGrey(u *url.URL) *types.HttpRequest {
	var existing types.HttpRequest
	if result := common.DB.First(&existing, "url = ? AND class = 'grey'", u.String()); result.RowsAffected <= 0 {
		return nil
	}

	existing.Count++
	existing.LastSeen = time.Now().UTC()
	common.DB.Save(existing)

	return &existing
}

func (svc *InnerHttpService) execute() {
}

func (svc *InnerHttpService) allItems(payload string) (interface{}, error) {
	var items []types.HttpRequest
	common.DB.Find(&items)
	return items, nil
}

func (svc *InnerHttpService) byFieldName(payload string) (interface{}, error) {
	var args types.ByNameRequest
	if err := json.Unmarshal([]byte(payload), &args); err != nil {
		svc.LogGeneric("error", "Marshalling request to JSON failed: %#v", err)
		return nil, err
	}

	var items []types.HttpRequest
	if result := common.DB.Where(map[string]interface{}{args.Name: args.Value}).Find(&items); result.Error != nil {
		svc.LogGeneric("error", "Failed to query database, error: %#v", result.Error)
		return nil, result.Error
	} else {
		fmt.Println("byFieldName items count: ", result.RowsAffected, args.Name, args.Value)
	}

	return items, nil
}

func (svc *InnerHttpService) update(payload string) (interface{}, error) {
	var item types.HttpRequest
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

func (svc *InnerHttpService) prune(payload string) (interface{}, error) {
	var greyitems, whiteitems, blackitems []types.HttpRequest

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
			re, _ := regexp.Compile(wv.MatchURL)
			if re.MatchString(gv.URL) {
				if result := common.DB.Delete(&gv); result.Error != nil {
					svc.LogGeneric("error", "Failed to delete item from database, error: %#v", result.Error)
					return nil, result.Error
				}
			}
		}
	}

	for _, bv := range blackitems {
		for _, gv := range greyitems {
			re, _ := regexp.Compile(bv.MatchURL)
			if re.MatchString(gv.URL) {
				if result := common.DB.Delete(&gv); result.Error != nil {
					svc.LogGeneric("error", "Failed to delete item from database, error: %#v", result.Error)
					return nil, result.Error
				}
			}
		}
	}

	return nil, nil
}
