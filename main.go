package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    capsolver_go "github.com/imvast/capsolver-go"
    "io/ioutil"
    "math/rand"
    "net/http"
    "os"
    "os/exec"
    "runtime"
    "sync"
    "time"
    logging "webshare-creator/internal"
)

var (
    Logger    = logging.Logger
    capSolver = capsolver_go.CapSolver{ApiKey: "CAI-83E66B6FCA907F945308B126B8DA05BD"}
)

type Webshare struct {
    Session *http.Client
}

func NewWebshare() *Webshare {
    return &Webshare{
        Session: &http.Client{},
    }
}

func (ws *Webshare) SolveCaptcha() (string, error) {
    for {
        s, err := capSolver.Solve(map[string]any{
            "type":        "ReCaptchaV2TaskProxyLess",
            "websiteURL":  "https://proxy2.webshare.io/",
            "websiteKey":  "6LeHZ6UUAAAAAKat_YS--O2tj_by3gv3r_l03j9d",
            "isInvisible": true,
        })
        if err != nil {
            continue
        }
        solutionStr := s.Solution.GRecaptchaResponse
        return solutionStr, nil
    }
}

func (ws *Webshare) Register(capKey string) (string, error) {
    url := "https://proxy.webshare.io/api/v2/register/"

    email := generateRandomEmail()
    payload := map[string]interface{}{
        "email":        email,
        "password":     "=vi*'*?s#\"bV2r7",
        "tos_accepted": true,
        "recaptcha":    capKey,
    }

    headers := map[string]string{
        "User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/112.0",
        "Accept":          "application/json, text/plain, */*",
        "Accept-Language": "en-US,en;q=0.5",
        "Accept-Encoding": "gzip, deflate, br",
        "Content-Type":    "application/json",
        "Connection":      "keep-alive",
        "Host":            "proxy.webshare.io",
        "Origin":          "https://proxy2.webshare.io",
        "Referer":         "https://proxy2.webshare.io/",
        "Sec-Fetch-Dest":  "empty",
        "Sec-Fetch-Mode":  "cors",
        "Sec-Fetch-Site":  "same-site",
        "Pragma":          "no-cache",
        "Cache-Control":   "no-cache",
        "TE":              "trailers",
    }

    jsonPayload, err := json.Marshal(payload)
    if err != nil {
        return "", err
    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
    if err != nil {
        return "", err
    }

    for key, value := range headers {
        req.Header.Set(key, value)
    }

    resp, err := ws.Session.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    respBody, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    var respJSON map[string]interface{}
    err = json.Unmarshal(respBody, &respJSON)
    if err != nil {
        return "", err
    }

    token, ok := respJSON["token"].(string)
    if !ok {
        return "", fmt.Errorf("failed to extract token from response")
    }

    return token, nil
}

func (ws *Webshare) GetProxy(authToken string) {
    url := "https://proxy.webshare.io/api/v2/proxy/config/"

    headers := map[string]string{
        "User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/112.0",
        "Accept":          "application/json, text/plain, */*",
        "Accept-Language": "en-US,en;q=0.5",
        "Accept-Encoding": "gzip, deflate, br",
        "Authorization":   fmt.Sprintf("Token %s", authToken),
        "Origin":          "https://proxy2.webshare.io",
        "Connection":      "keep-alive",
        "Referer":         "https://proxy2.webshare.io/",
        "Sec-Fetch-Dest":  "empty",
        "Sec-Fetch-Mode":  "cors",
        "Sec-Fetch-Site":  "same-site",
        "Pragma":          "no-cache",
        "Cache-Control":   "no-cache",
        "TE":              "trailers",
    }

    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        fmt.Println(err)
        return
    }

    for key, value := range headers {
        req.Header.Set(key, value)
    }

    resp, err := ws.Session.Do(req)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer resp.Body.Close()

    respBody, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println(err)
        return
    }

    var respJSON map[string]interface{}
    err = json.Unmarshal(respBody, &respJSON)
    if err != nil {
        fmt.Println(err)
        return
    }

    user, _ := respJSON["username"].(string)
    pasw, _ := respJSON["password"].(string)
    proxy := fmt.Sprintf("%s-rotate:%s@p.webshare.io:80", user, pasw)

    file, err := os.OpenFile("./proxies.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer file.Close()

    file.WriteString(proxy + "\n")
    Logger.Info().Str("proxy", proxy).Msg("Created")
}

func generateRandomEmail() string {
    rand.Seed(time.Now().UnixNano())

    chars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
    email := make([]rune, 8)
    for i := range email {
        email[i] = chars[rand.Intn(len(chars))]
    }

    return fmt.Sprintf("%s@outlook.fr", string(email))
}

func main() {
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)

        go func() {
            defer wg.Done()

            ws := NewWebshare()
            capKey, err := ws.SolveCaptcha()
            if err != nil {
                fmt.Println(err)
                return
            }

            authToken, err := ws.Register(capKey)
            if err != nil {
                fmt.Println(err)
                return
            }

            ws.GetProxy(authToken)
        }()
    }

    wg.Wait()
}

func clearScreen() {
    switch runtime.GOOS {
    case "windows":
        cmd := exec.Command("cmd", "/c", "cls")
        cmd.Stdout = os.Stdout
        cmd.Run()
    default:
        cmd := exec.Command("clear")
        cmd.Stdout = os.Stdout
        cmd.Run()
    }
}

func init() {
    clearScreen()
    blue := "\033[36m"
    reset := "\033[0m"
    exaa := (`
  ╔═╗┌─┐╔═╗┬ ┬┌─┐┬─┐┌─┐
  ║ ╦│ │╚═╗├─┤├─┤├┬┘├┤
  ╚═╝└─┘╚═╝┴ ┴┴ ┴┴└─└─┘
`)
    fmt.Println(blue + exaa + reset)
}
