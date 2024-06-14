package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "time"

    "github.com/mattermost/mattermost/server/public/plugin"
)

type Plugin struct {
    plugin.MattermostPlugin
}

type HostInfo struct {
    HostType   string    `json:"host_type"`
    IPAddress  string    `json:"ip_address"`
    OnlineTime time.Time `json:"online_time"`
}

func (p *Plugin) OnActivate() error {
    p.API.LogInfo("Brute Ratel C4 Host Notifier Plugin activated")
    return nil
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/notify" && r.Method == http.MethodPost {
        body, err := ioutil.ReadAll(r.Body)
        if err != nil {
            p.API.LogError("Error reading request body", "err", err)
            http.Error(w, "Error reading request body", http.StatusInternalServerError)
            return
        }

        var hostInfo HostInfo
        err = json.Unmarshal(body, &hostInfo)
        if err != nil {
            p.API.LogError("Error unmarshalling request body", "err", err)
            http.Error(w, "Error unmarshalling request body", http.StatusInternalServerError)
            return
        }

        p.sendHostInfoToMattermost(hostInfo)
    }
}

func (p *Plugin) sendHostInfoToMattermost(info HostInfo) {
    webhookURL := p.API.GetConfig().PluginSettings.Plugins["brc4_mattermost_plugin"].(map[string]interface{})["WebhookURL"].(string)
    message := fmt.Sprintf("Host Type: %s\nIP Address: %s\nOnline Time: %s", info.HostType, info.IPAddress, info.OnlineTime.Format(time.RFC3339))

    payload := map[string]string{"text": message}
    payloadBytes, err := json.Marshal(payload)
    if err != nil {
        p.API.LogError("Error marshalling payload", "err", err)
        return
    }

    resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(payloadBytes))
    if err != nil {
        p.API.LogError("Error sending message to Mattermost", "err", err)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := ioutil.ReadAll(resp.Body)
        p.API.LogError("Error from Mattermost webhook", "status", resp.StatusCode, "body", string(body))
    }
}

func main() {
    plugin.ClientMain(&Plugin{})
}

