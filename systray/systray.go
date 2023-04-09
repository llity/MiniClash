package systray

import (
	"fmt"
	"os"
	"io"
	"runtime"
	"strconv"
	"time"
	"net/http"
 	 
	"github.com/atotto/clipboard"
  "gopkg.in/yaml.v3"
 
 	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"

	"github.com/Dreamacro/clash/constant"
	proxy "github.com/Dreamacro/clash/listener"
	"github.com/Dreamacro/clash/tunnel"

	"github.com/gfw-list/TrayedClash/icon"
	"github.com/gfw-list/TrayedClash/sysproxy"
)

func init() {
	if runtime.GOOS == "windows" {
		currentDir, _ := os.Getwd()
		constant.SetHomeDir(currentDir)
	}

	go func() {
		runtime.LockOSThread()
		systray.Run(onReady, onExit)
		runtime.UnlockOSThread()
	}()
}
 
func addNewConfig() { 
  currentDir, _ := os.Getwd()
  configUrl , err := clipboard.ReadAll()

  client := &http.Client{Timeout: 10 * time.Second}
  req, err := http.NewRequest("GET", configUrl, nil)
  if err != nil {
          return
  }
 	req.Header.Add("User-Agent", "clash")
 
  // Get the data
  resp, err := client.Do(req)
  
  defer resp.Body.Close()

  configFile, _ := io.ReadAll(resp.Body)
  
  //validate if the resp cont is yaml or not
  var yml interface{} 
  if err := yaml.Unmarshal([]byte(configFile), &yml); err != nil {
 	   return
  }
  os.WriteFile(currentDir + `/` + "config.yaml", configFile, 0644)
  if err != nil {
    panic("Unable to write data into the file")
  }
}

func onReady() {
	systray.SetIcon(icon.Data)
	systray.SetTitle("Clash")
	systray.SetTooltip("Mini Clash")

	mTitle := systray.AddMenuItem("Mini Clash", "")
	systray.AddSeparator()

	mGlobal := systray.AddMenuItem("Global", "Set as Global")
	mRule := systray.AddMenuItem("Rule", "Set as Rule")
	mDirect := systray.AddMenuItem("Direct", "Set as Direct")
	systray.AddSeparator()

	mEnabled := systray.AddMenuItem("Set as System Proxy", "Turn on/off Proxy")
	mURL := systray.AddMenuItem("Open Dashboard", "Open Clash Dashboard")
	systray.AddSeparator()

  mNewConfig := systray.AddMenuItem("Add from Copied URL", "Subscribe")
  mConfigFolder := systray.AddMenuItem("Open Config Folder", "Open Profile Folder")
	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Exit", "Quit Clash")

	go func() {
		t := time.NewTicker(time.Duration(time.Second))
		defer t.Stop()

		SavedPort := proxy.GetPorts().Port
		for {
			<-t.C

			switch tunnel.Mode() {
			case tunnel.Global:
				if mGlobal.Checked() {
				} else {
					mGlobal.Check()
					mRule.Uncheck()
					mDirect.Uncheck()
				}
			case tunnel.Rule:
				if mRule.Checked() {
				} else {
					mGlobal.Uncheck()
					mRule.Check()
					mDirect.Uncheck()
				}
			case tunnel.Direct:
				if mDirect.Checked() {
				} else {
					mGlobal.Uncheck()
					mRule.Uncheck()
					mDirect.Check()
				}
			}

			if mEnabled.Checked() {
				p := proxy.GetPorts().MixedPort
				if SavedPort != p {
					SavedPort = p
					err := sysproxy.SetSystemProxy(
						&sysproxy.ProxyConfig{
							Enable: true,
							Server: "127.0.0.1:" + strconv.Itoa(SavedPort),
						})
					if err != nil {
						continue
					}
				}
			}

			p, err := sysproxy.GetCurrentProxy()
			if err != nil {
				continue
			}

			if p.Enable && p.Server == "127.0.0.1:"+strconv.Itoa(SavedPort) {
				if mEnabled.Checked() {
				} else {
					mEnabled.Check()
				}
			} else {
				if mEnabled.Checked() {
					mEnabled.Uncheck()
				} else {
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case <-mTitle.ClickedCh:
				fmt.Println("Title Clicked")
			case <-mGlobal.ClickedCh:
				tunnel.SetMode(tunnel.Global)
			case <-mRule.ClickedCh:
				tunnel.SetMode(tunnel.Rule)
			case <-mDirect.ClickedCh:
				tunnel.SetMode(tunnel.Direct)
			case <-mEnabled.ClickedCh:			
				if mEnabled.Checked() {
					err := sysproxy.SetSystemProxy(sysproxy.GetSavedProxy())
					if err != nil {
					} else {
						mEnabled.Uncheck()
					}
				} else {
					err := sysproxy.SetSystemProxy(
						&sysproxy.ProxyConfig{
							Enable: true,
							Server: "127.0.0.1:" + strconv.Itoa(proxy.GetPorts().MixedPort),
						})
					if err != nil {
					} else {
						mEnabled.Check()
					}
				}
			case <-mURL.ClickedCh:
				open.Run("http://127.0.0.1:8090/")
			case <-mNewConfig.ClickedCh:
				addNewConfig() 	
			case <-mConfigFolder.ClickedCh:	
			  currentDir, _ := os.Getwd()			
			  open.Run(currentDir + `/`)
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	for {
		err := sysproxy.SetSystemProxy(sysproxy.GetSavedProxy())
		if err != nil {
			continue
		} else {
			break
		}
	}

	os.Exit(1)
}
