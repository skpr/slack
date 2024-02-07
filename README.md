Slack
=====

A library for posting consistent Slack messages across the Skpr platform.

## Usage


```go
client, err := slack.NewClient([]string{
    "https://hooks.slack.com/services/xxxxxxxxx/yyyyyyyyy/zzzzzzzzz",
})
if err != nil {
    panic(err)
}

params := slack.PostMessageParams{
    Icon: "https://raw.githubusercontent.com/skpr/slack/main/icons/application_drupal.png",
    Context: map[string]string{
        "Account":    "123456789",
        "Anomaly ID": "xxxxxx-yyyyyy-zzzzzz",
    },
    Description: "This is the description field",
    Reason:      "This is the reason field.",
    Dashboard:   "https://skpr.com.au",
}

err = client.PostMessage(params)
if err != nil {
    panic(err)
}
```
