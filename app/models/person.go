package models

type Person struct {
  LocalGuid    string
  RemoteGuid   string
  DisplayName  string
  ImageSrc     string
  PodUrl       string
  ProfileUrl   string
  Email        string
  RSAPubKey    string
  AccountIdentifier string
}

// Connect initiates and synchronously
// performs a sharing notification and
// datastore write
func (self *Person) Connect(user *User) (error) {
  // XXX validation?
//  if(self.IsLocal() == false) {
    result, err := SendSharingNotification(user, self)
    if err != nil {
      panic(err)
    }
    if result.StatusCode == 200 || result.StatusCode == 202 {
     panic("implement datastore write")
    } else {
      panic("received: "+result.Status)
    }
 // }
  return nil
}

func (self *Person) IsLocal() (bool) {
  return (self.LocalGuid == self.RemoteGuid)
}
