package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/emptyinterface/window"
	"golang.org/x/net/websocket"
)

var (
	host     = flag.String("host", ":4443", "host to serve https on")
	cert_key = flag.String("cert_key", "cert/key.pem", "keys for https")
	cert     = flag.String("cert", "cert/cert.pem", "keys for https")
	ssh_keys = flag.String("ssh_keys", "$HOME/.ssh", "ssh keys to servers")

	concurrency       = flag.Int("concurrency", 40, "how many ssh/api requests allowed concurrently")
	rate              = flag.Int("rate", 300, "how many ssh/api calls can be made within a given period (rate_interval)")
	rate_interval     = flag.Duration("rate_interval", time.Second, "the duration constraint to the ssh/api call rate")
	region_interval   = flag.Duration("region_interval", 5*time.Minute, "polling interval for aws service information")
	instance_interval = flag.Duration("instance_interval", 60*time.Second, "polling interval for instance sysinfo stats")

	dev = flag.Bool("dev", false, "recompile templates on refresh")
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	if _, exists := os.LookupEnv("AWS_REGION"); !exists {
		log.Fatal("Must load AWS_* environment variables")
	}
}

func main() {

	flag.Parse()

	region := window.NewRegion(os.Getenv("AWS_REGION"))
	region.Throttle = window.NewThrottle(*concurrency, *rate, *rate_interval)
	region.SetSSHKeyPath(os.ExpandEnv(*ssh_keys))

	var (
		templateSet = NewTemplateSet()

		pub = NewPublisher(func(path string) string {
			region.Lock()
			defer region.Unlock()
			if *dev {
				templateSet.Build()
			}
			start := time.Now()
			defer func() { fmt.Println(path, "template executed in", time.Since(start)) }()
			switch {
			case path == "/data/":
				return templateSet.Execute("_region.html", region)
			case path == "/data/classic":
				return templateSet.Execute("_classic.html", region.Classic)
			case strings.HasPrefix(path, "/data/vpc/"):
				vpcname := strings.TrimPrefix(path, "/data/vpc/")
				var route string
				if i := strings.IndexByte(vpcname, '/'); i > 0 {
					vpcname, route = vpcname[:i], strings.Replace(vpcname[i+1:], "/", "_", -1)
				}
				for _, vpc := range region.VPCs {
					if vpc.Name == vpcname {
						if len(route) > 0 {
							templateName := "_" + route + ".html"
							if templateSet.templates.Lookup(templateName) != nil {
								return templateSet.Execute(templateName, vpc)
							}
							return "bad template"
						} else {
							return templateSet.Execute("_vpc.html", vpc)
						}
					}
				}
				return fmt.Sprintf("%q vpc not found", vpcname)

			case strings.HasPrefix(path, "/data/"):
				templateName := "_" + strings.TrimPrefix(path, "/data/") + ".html"
				if templateSet.templates.Lookup(templateName) != nil {
					return templateSet.Execute(templateName, region)
				}
				return "bad template"
			default:
				return "stfu"
			}
		})
	)

	start := time.Now()
	if err := region.Refresh(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("First refresh in", time.Since(start))

	go func() {
		region_ticker := time.NewTicker(*region_interval)
		instance_ticker := time.NewTicker(*instance_interval)
		defer region_ticker.Stop()
		defer instance_ticker.Stop()
		for {
			func() {
				// defer func() {
				// 	if e := recover(); e != nil {
				// 		log.Printf("panic: %v", e)
				// 	}
				// }()
				select {
				case <-region_ticker.C:
					if err := region.Refresh(); err != nil {
						log.Println(err)
					}
				case <-instance_ticker.C:
					for _, errchans := range region.RefreshInstances() {
						for _, errchan := range errchans {
							if err := <-errchan; err != nil {
								fmt.Println(err)
							}
						}
					}
				}
				pub.Publish()
			}()
		}
	}()

	mux := NewMux(func() {
		if *dev {
			start := time.Now()
			templateSet.Build()
			fmt.Println("templates recompiled in", time.Since(start))
		}
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if err := templateSet.templates.ExecuteTemplate(w, "index.html", struct {
			Request *http.Request
			Region  *window.Region
		}{
			Request: req,
			Region:  region,
		}); err != nil {
			log.Println(err)
		}
	})
	mux.HandleFunc("/data/", func(w http.ResponseWriter, req *http.Request) {
		websocket.Handler(func(ws *websocket.Conn) {

			defer ws.Close()

			publish := make(chan string, 1) // so we don't miss it
			pub.Subscribe(req.URL.Path, publish)
			defer pub.Unsubscribe(req.URL.Path, publish)

			for {
				if err := websocket.Message.Send(ws, <-publish); err != nil {
					if !strings.Contains(err.Error(), "broken pipe") {
						fmt.Println("websocket send derr:", err)
					}
					return
				}
			}

		}).ServeHTTP(w, req)
	})

	mux.HandleFunc("/_data/", func(w http.ResponseWriter, req *http.Request) {
		parts := strings.SplitN(strings.TrimPrefix(req.URL.Path, "/_data/"), "/", 2)
		if len(parts) < 2 {
			http.Error(w, "invalid _data path", http.StatusBadRequest)
			return
		}
		if v, exists := region.Items[parts[1]]; exists {
			fmt.Fprint(w, templateSet.Execute("_"+parts[0]+"_data.html", v))
			return
		}
		http.Error(w, fmt.Sprintf("%s/%s not found", parts[0], parts[1]), http.StatusNotFound)
	})

	mux.Handle("/assets/", http.FileServer(http.Dir("./web/")))

	fmt.Println("ready")
	log.Fatal(http.ListenAndServeTLS(*host, *cert, *cert_key, mux))

}
