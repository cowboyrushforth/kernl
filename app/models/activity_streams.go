package models

type MediaLink struct {
  Duration int64 `json:"duration"`
  Height int `json:"width"`
  Width int `json:"height"`
  Url string `json:"string"`
}

type ActivityObject struct {
  Attachments []string `json:"attachments,omitempty"`
  Author *ActivityObject `json:"author,omitempty"`
  Content string `json:"content"`
  DisplayName string `json:"displayName"`
  DownstreamDuplicates []string `json:"downstreamDuplicates,omitempty"`
  Id string `json:"id"`
  Image *MediaLink `json:"image,omitempty"`
  ObjectType string `json:"objectType"`
  Published string `json:"published"`
  Summary string `json:"summary"`
  UpdatedAt string `json:"updated"`
  UpstreamDuplicates []string `json:"upstreamDuplicates,omitempty"`
  Url string `json:"url"`
}

type Activity struct {
  Actor ActivityObject `json:"actor"`
  Content string `json:"content"`
  Generator ActivityObject `json:"generator"`
  Icon MediaLink `json:"icon"`
  Id string `json:"id"`
  Object ActivityObject `json:"object"`
  Published string `json:"published"`
  Provider ActivityObject `json:"provider"`
  Target ActivityObject `json:"target"`
  Title string `json:"title"`
  UpdatedAt string `json:"updated"`
  Url string `json:"url"`
  Verb string `json:"verb"`
}

type Collection struct {
  TotalItems int `json:"totalItems"`
  Items []ActivityObject `json:"items"`
  Url string `json:"url"`
}
