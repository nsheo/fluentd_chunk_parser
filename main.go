package main

import (
    "fmt"
    "os"
    "io/ioutil"
    "bufio"
    "bytes"
    "github.com/tinylib/msgp/msgp"
    "encoding/json"
    "strings"
    "github.com/namsral/flag"
    "encoding/base64"
    "net/http"
)

type Settings struct {
    Settings []Setting `json:"settings"`
}
 
type Setting struct {
    Filename       string     `json:"file_name"`
    DecodeTarget   []string   `json:"base64_decode_target"`
    SendTarget     string     `json:"send_target"`
    Targetpath     string     `json:"target_path"`
}

func UnMarshal(filename string) []byte {
    file, err := os.Open(filename)

    if err != nil {
        panic(err)
    }
    defer file.Close()

    bufr := bufio.NewReader(file)

    var js bytes.Buffer

    _, err2 := msgp.CopyToJSON(&js, bufr)
    if err2 != nil {
        panic(fmt.Sprintf("Cannot convert MessagePack to JSON: %v", err2))
    }

    return js.Bytes() 
}

func Find(slice []string, val string) (bool) {
    for _, item := range slice {
        if item == val {
            return true
        }
    }
    return false
}


func RunItem(setting Setting){
    decodeTarget := setting.DecodeTarget
    parseData := UnMarshal(setting.Filename)
    
    bufArr := strings.Split(string(parseData), "]")
    for _, strBuffer := range bufArr {
        if len(strBuffer) == 0 {
            break
        }
        var bufTarget = strBuffer + "]"
        var arr []interface{}
        _ = json.Unmarshal([]byte(bufTarget), &arr)
        var bufMap map[string]interface{}
        bufMap = arr[2].(map[string]interface{})
        if len(decodeTarget) > 0 {
            for k, v := range bufMap {
                if Find(decodeTarget, k) {
                    sDec, _ := base64.StdEncoding.DecodeString(v.(string))
                    bufMap[k] = string(sDec)
                }
            }
        }
        OutputData(setting, bufMap, arr[0].(string), arr[1].(float64))
    }
}

func OutputData(setting Setting, data map[string] interface{}, tag string, time float64) {
    timeStr := fmt.Sprintf("%f", time)
    var jsonBuffer = bytes.NewBuffer([]byte{})
    jsonEncoder := json.NewEncoder(jsonBuffer)
    jsonEncoder.SetEscapeHTML(false)
    jsonEncoder.Encode(data)

    if setting.SendTarget == "file" {
        _ = ioutil.WriteFile(setting.Targetpath + tag + "_" + timeStr + ".json", jsonBuffer.Bytes(), 0644)
    } else if setting.SendTarget == "fluentd" {
        //Send To Fluentd HTTP
        var httpTarget = "http://127.0.0.1:9880/"
        if setting.Targetpath != "" {
            httpTarget = "http://"+setting.Targetpath+"/"
        }
        httpTarget = httpTarget +tag +"?time="+timeStr
        //httpTarget = httpTarget +"htest.logs?time="+timeStr
        HTTPSender(httpTarget, jsonBuffer)
    } else if setting.SendTarget == "stdout" {
        fmt.Println("Tag: %s, Time: %s, Record: %s", tag, timeStr, jsonBuffer.String())
    } else {
        fmt.Println("Tag: %s, Time: %s, Record: %s", tag, timeStr, jsonBuffer.String())
    }
}

func HTTPSender(url string, b *bytes.Buffer) {
    req, _ := http.NewRequest("POST", url, b)
    client := &http.Client{}
    res, httpError := client.Do(req)
    if httpError != nil {
        fmt.Sprintf("HTTP Sending Error : %v", httpError)
    }
    defer res.Body.Close()
    fmt.Println("Fluentd HTTP response Status:", res.Status)
}

func main(){
    settingsFile := flag.String("settings", "./settings.json", "files with settings")
    flag.Parse()

    settings := Settings{}
   
    if fd, err := os.Open(*settingsFile); err == nil {
        dec := json.NewDecoder(fd)
        if err = dec.Decode(&settings); err != nil {
            panic(fmt.Sprintf("Setting Reading Error : %v", err))
        }
    } else {
        panic(fmt.Sprintf("Setting Reading Error : %v", err))
    }
    
    for i := 0; i < len(settings.Settings); i++ {
        RunItem(settings.Settings[i])
    }
}
