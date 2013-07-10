package models


type Activity struct {
}

type MediaLink struct {
  Duration int64 `json:"duration"`
  Height int `json:"width"`
  Width int `json:"height"`
  Url string `json:"string"`
}


type ActivityObject struct {
  Attachments []string `json:"attachments"`
//  Author ActivityObject `json:"author"`
  Content string `json:"content"`
  DisplayName string `json:"displayName"`
  DownstreamDuplicates []string `json:"downstreamDuplicates"`
  Id string `json:"id"`
  Image string `json:"image"`
  ObjectType string `json:"objectType"`
  Published int64 `json:"published"`
  Summary string `json:"summary"`
  Updated int64 `json:"updated"`
  UpstreamDuplicates []string `json:"upstreamDuplicates"`
  Url string `json:"url"`
}

type Collection struct {
  TotalItems int `json:"totalItems"`
  Items []ActivityObject `json:"items"`
  Url string `json:"url"`
}
