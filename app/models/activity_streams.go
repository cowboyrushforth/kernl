package models

type MediaLink struct {
  Duration int64 `json:"duration"`
  Height int `json:"width"`
  Width int `json:"height"`
  Url string `json:"string"`
}

type InnerLink struct {
  Href string `json:"href,omitempty"`
  Url string `json:"url,omitempty"`
  TotalItems *int `json:"totalItems,omitempty"`
}

type Links struct {
  Self *InnerLink `json:"self,omitempty"`
  ActivityInbox *InnerLink `json:"activity-inbox,omitempty"`
  ActivityOutbox *InnerLink `json:"activity-outbox,omitempty"`
}

type PumpIo struct {
  Shared bool `json:"shared"`
  Followed bool `json:"followed"`
}

type ActivityObject struct {
  Attachments []string `json:"attachments,omitempty"`
  Author *ActivityObject `json:"author,omitempty"`
  Content string `json:"content,omitempty"`
  DisplayName string `json:"displayName"`
  PreferredUsername string `json:"preferredUsername"`
  DownstreamDuplicates []string `json:"downstreamDuplicates,omitempty"`
  Id string `json:"id"`
  Image *MediaLink `json:"image,omitempty"`
  ObjectType string `json:"objectType"`
  Published string `json:"published,omitempty"`
  Summary string `json:"summary,omitempty"`
  UpdatedAt string `json:"updated"`
  UpstreamDuplicates []string `json:"upstreamDuplicates,omitempty"`
  Url string `json:"url"`
  Links *Links `json:"links,omitempty"`
  Favorites *InnerLink `json:"favorites,omitempty"`
  Lists *InnerLink `json:"lists,omitempty"`
  Followers *InnerLink `json:"followers,omitempty"`
  Following *InnerLink `json:"following,omitempty"`
  Liked bool `json:"liked"`
  PumpIo PumpIo `json:"pump_io"`
}

type Activity struct {
  Actor *ActivityObject `json:"actor"`
  Content string `json:"content"`
  Generator *ActivityObject `json:"generator"`
  Icon MediaLink `json:"icon"`
  Id string `json:"id"`
  Object *ActivityObject `json:"object"`
  Published string `json:"published"`
  Provider *ActivityObject `json:"provider"`
  Target *ActivityObject `json:"target"`
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
