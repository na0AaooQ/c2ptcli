package main

import (
  "fmt"
  "os"
  "net/http"       // チャットAPIコール用
  "net/url"        // チャットAPIコール用
  "strings"        // チャットAPIコール用
  "bytes"          // LINE Messaging APIコール用
  "encoding/json"  // JSONパース用
  "io/ioutil"      // JSONパース用
)

// Slack APIメッセージ投稿結果レスポンス(json)格納用構造体
// Slackは存在しないchannelにメッセージを投稿した場合でも、ステータスコード200を返す為、Slack APIレスポンスのok,errorをチェックする
type SlackResponse struct {
  Ok    bool   `json:"ok"`
  Error string `json:"error"`
}

// LINE Messaging API認証レスポンス(json)格納用構造体
type LineAuthResponse struct {
  AccessToken   string `json:"access_token"`
  ExpireIn      int32  `json:"expires_in"`
  TokenType     string `json:"token_type"`
}

///////////////////////////////////////////////////////////////////////////
// Slackへメッセージを投稿する
///////////////////////////////////////////////////////////////////////////
func SlackMessagePost(chatKind, apiUrl, apiToken, channel, message string) error {
  values := url.Values{}
  values.Set("token", apiToken)
  values.Add("channel", channel)
  values.Add("text", message)

  req, err := http.NewRequest(
    "POST",
    apiUrl,
    strings.NewReader(values.Encode()),
  )
  if err != nil {
    fmt.Println(err)
    return err
  }

  // Content-Type 設定
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

  // チャットへメッセージを投稿
  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    fmt.Println(err)
    return err
  }
  defer resp.Body.Close()

  // メッセージ投稿結果の表示
  if resp.StatusCode == http.StatusOK {
 
    // Slack APIレスポンスをパースする
    slackRespArray, _ := ioutil.ReadAll(resp.Body)
    slackRespJsonBytes := ([]byte)(slackRespArray)
    slackRespData := new(SlackResponse)

    // Slack APIレスポンスのok:がfalseの場合、メッセージ投稿エラー
    if err := json.Unmarshal(slackRespJsonBytes, slackRespData); err != nil {
      fmt.Println("JSON Unmarshal error:", err)
    }
    if slackRespData.Ok == false {
      fmt.Println(chatKind + "メッセージ投稿に失敗しました。")
      fmt.Println("  " + chatKind + "メッセージ投稿結果=[", slackRespData.Ok, "] メッセージ投稿エラーメッセージ=[", slackRespData.Error, "]")
    } else {
      fmt.Println(chatKind + "メッセージ投稿に成功しました。")
    }

  } else {
    fmt.Println(chatKind + "メッセージ投稿に失敗しました。")
  }
  fmt.Println("  " + chatKind + "メッセージ投稿ステータスコード=[", resp.StatusCode, "] レスポンス内容=[", resp.Status, "]")

  return err
}

///////////////////////////////////////////////////////////////////////////
// LINEへメッセージを投稿する
///////////////////////////////////////////////////////////////////////////
func LineMessagePost(chatKind, apiAuthUrl, apiPushMessageUrl, lineChannelId, lineChannelSecret, lineUid, message string) error {

  values := url.Values{}
  values.Set("client_id", lineChannelId)
  values.Add("client_secret", lineChannelSecret)
  values.Add("grant_type", "client_credentials")

  req, err := http.NewRequest(
    "POST",
    apiAuthUrl,
    strings.NewReader(values.Encode()),
  )
  if err != nil {
    fmt.Println(err)
    return err
  }

  // Content-Type 設定
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

  // LINE Messaging APIのアクセストークン取得
  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    fmt.Println(err)
    return err
  }
  defer resp.Body.Close()

  // メッセージ投稿結果の表示
  if resp.StatusCode == http.StatusOK {
 
    // LINE Messaging APIレスポンスをパースする
    lineRespArray, _ := ioutil.ReadAll(resp.Body)
    lineRespJsonBytes := ([]byte)(lineRespArray)
    lineRespData := new(LineAuthResponse)

    // LINE Messaging APIアクセストークン取得結果をチェック
    if err := json.Unmarshal(lineRespJsonBytes, lineRespData); err != nil {
      fmt.Println("JSON Unmarshal error:", err)
    }
    if len(lineRespData.AccessToken) < 1 {
      fmt.Println(chatKind + " APIアクセストークン取得に失敗しました。")
      fmt.Println("  " + chatKind + " APIアクセストークン取得結果=[", lineRespData.AccessToken, "] アクセストークン取得エラーメッセージ=[", lineRespData.ExpireIn, "]")
    } else {
      fmt.Println(chatKind + " APIアクセストークン取得に成功しました。")

      LinePushMessagePost(chatKind, apiPushMessageUrl, lineRespData.AccessToken, lineUid, message)

    }

  } else {
    fmt.Println(chatKind + " APIアクセストークン取得に失敗しました。")
  }
  fmt.Println("  " + chatKind + " APIアクセストークン取得ステータスコード=[", resp.StatusCode, "] レスポンス内容=[", resp.Status, "]")

  return err
}

///////////////////////////////////////////////////////////////////////////
// LINEへプッシュメッセージ(プッシュメッセージとはLINEメッセージの一種)を投稿する
///////////////////////////////////////////////////////////////////////////
func LinePushMessagePost(chatKind, apiPushMessageUrl, lineAccessToken, lineUid, message string) error {

  // LINEへ投稿するプッシュメッセージリクエストjsonをセット
  jsonStr := `{
    "to": "` + lineUid + `",
    "messages": [
      {
        "type": "text",
        "text": "` + message + `"
      }
    ]
  }`

  req, err := http.NewRequest(
    "POST",
    apiPushMessageUrl,
    bytes.NewBuffer([]byte(jsonStr)),
  )
  if err != nil {
    fmt.Println(err)
    return err
  }

  // Content-Type 設定
  req.Header.Set("Content-Type", "application/json")
  req.Header.Set("Authorization", "Bearer " + lineAccessToken)

  // チャットへメッセージ投稿
  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    return err
  }

  defer resp.Body.Close()

  // メッセージ投稿結果の表示
  if resp.StatusCode == http.StatusOK {
    fmt.Println(chatKind + "メッセージ投稿に成功しました。")
  } else {
    fmt.Println(chatKind + "メッセージ投稿に失敗しました。")
  }
  fmt.Println("  " + chatKind + "メッセージ投稿ステータスコード=[", resp.StatusCode, "] レスポンス内容=[", resp.Status, "]")
  return err

}

///////////////////////////////////////////////////////////////////////////
// Chatworkへメッセージを投稿する
///////////////////////////////////////////////////////////////////////////
func ChatworkMessagePost(chatKind, apiUrl, apiToken, message string) error {
  values := url.Values{}
  values.Set("body", message)

  req, err := http.NewRequest(
    "POST",
    apiUrl,
    strings.NewReader(values.Encode()),
  )
  if err != nil {
    return err
  }

  // Content-Type 設定
  req.Header.Set("X-ChatWorkToken", apiToken)
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

  // チャットへメッセージ投稿
  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    return err
  }

  defer resp.Body.Close()

  // メッセージ投稿結果の表示
  if resp.StatusCode == http.StatusOK {
    fmt.Println(chatKind + "メッセージ投稿に成功しました。")
  } else {
    fmt.Println(chatKind + "メッセージ投稿に失敗しました。")
  }
  fmt.Println("  " + chatKind + "メッセージ投稿ステータスコード=[", resp.StatusCode, "] レスポンス内容=[", resp.Status, "]")

  return err
}

///////////////////////////////////////////////////////////////////////////
// コマンドライン入力パラメータをチェックする
///////////////////////////////////////////////////////////////////////////
func CheckCliInputParameter() {

//  fmt.Println(os.Args)

  if len(os.Args) != 4 {
    fmt.Println("指定された引数の数が間違っています。")
    fmt.Println("Usage:")

    fmt.Println("[slackへメッセージを投稿する場合]")
    fmt.Println("  export SLACK_API_TOKEN=\"slackのAPIトークンをセットする\"")
    fmt.Println("  ./c2ptcli slack メッセージ投稿先のチャネル名(例:general) \"slackへ投稿したいメッセージ\"\n")

    fmt.Println("[chatworkへメッセージを投稿する場合]")
    fmt.Println("  export CHATWORK_API_TOKEN=\"chatworkのAPIトークンをセットする\"")
    fmt.Println("  ./c2ptcli chatwork メッセージ投稿先のルーム番号 \"chatworkへ投稿したいメッセージ\"\n")

    fmt.Println("[LINEへメッセージを投稿する場合]")
    fmt.Println("  export LINE_CHANNEL_ID=\"LINE Messaging APIのChannel ID文字列をセットする\"")
    fmt.Println("  export LINE_CHANNEL_SECRET=\"LINE Messaging APIのChannel Secret文字列列をセットする\"")
    fmt.Println("  ./c2ptcli line メッセージ投稿先のUID(LINE User Id) \"LINEへ投稿したいプッシュメッセージ\"\n")

    os.Exit(1)
  }

}
///////////////////////////////////////////////////////////////////////////
// main
///////////////////////////////////////////////////////////////////////////
func main() {

  // 入力パラメータのチェック
  CheckCliInputParameter()

  // メッセージを投稿するチャットの種類を取得
  chatKind := os.Args[1]

  // チャットの種類により、メッセージ投稿用のパラメータを生成
  switch chatKind {

    case "slack":

      postChannel := os.Args[2]
      postMessage := os.Args[3]

      apiUrl := "https://slack.com/api/chat.postMessage"
      if len(os.Getenv("SLACK_API_TOKEN")) != 0 {

        apiToken := os.Getenv("SLACK_API_TOKEN")
        SlackMessagePost(chatKind, apiUrl, apiToken, postChannel, postMessage)

      } else {

        fmt.Println(chatKind + "メッセージ投稿用のAPIトークンを環境変数 SLACK_API_TOKEN にセットして下さい。")
        fmt.Println("Linux,Mac系の環境変数設定方法(bashの例):")
        fmt.Println("  export SLACK_API_TOKEN=\"APIトークン文字列\"")
        os.Exit(1)

      }

    case "chatwork":

      postChannel := os.Args[2]
      postMessage := os.Args[3]

      apiUrl := "https://api.chatwork.com/v2/rooms/" + postChannel + "/messages"
      if len(os.Getenv("CHATWORK_API_TOKEN")) != 0 {

        apiToken := os.Getenv("CHATWORK_API_TOKEN")
        ChatworkMessagePost(chatKind, apiUrl, apiToken, postMessage)

      } else {

        fmt.Println(chatKind + "メッセージ投稿用のAPIトークンを環境変数 CHATWORK_API_TOKEN にセットして下さい。")
        fmt.Println("Linux,Mac系の環境変数設定方法(bashの例):")
        fmt.Println("  export CHATWORK_API_TOKEN=\"APIトークン文字列\"")
        os.Exit(1)

      }

    case "line":

      postLineUid := os.Args[2]
      postMessage := os.Args[3]

      apiAuthUrl := "https://api.line.me/v2/oauth/accessToken"
      apiPushMessageUrl := "https://api.line.me/v2/bot/message/push"
      if len(os.Getenv("LINE_CHANNEL_ID")) != 0 && len(os.Getenv("LINE_CHANNEL_SECRET")) != 0 {

        lineChannelId := os.Getenv("LINE_CHANNEL_ID")
        lineChannelSecret := os.Getenv("LINE_CHANNEL_SECRET")
        LineMessagePost(chatKind, apiAuthUrl, apiPushMessageUrl, lineChannelId, lineChannelSecret, postLineUid, postMessage)

      } else {

        fmt.Println(chatKind + "メッセージ投稿用のAPIトークンを環境変数 LINE_CHANNEL_IDとLINE_CHANNEL_SECRET にセットして下さい。")
        fmt.Println("Linux,Mac系の環境変数設定方法(bashの例):")
        fmt.Println("  export LINE_CHANNEL_ID=\"LINE Messaging APIのChannel ID文字列\"")
        fmt.Println("  export LINE_CHANNEL_SECRET=\"LINE Messaging APIのChannel Secret文字列\"")
        os.Exit(1)

      }

    default:
      fmt.Printf("未対応のチャットツールが指定されました。指定されたチャットの種類 = [%s]\n", chatKind)
    
  }

}
