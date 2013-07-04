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
  if(self.IsLocal() == false) {
    err := SendSharingNotification(user, self)
    if err != nil {
      panic(err)
    }
    panic("implement datastore write")
  }
  return nil
}

func (self *Person) IsLocal() (bool) {
  return (self.LocalGuid == self.RemoteGuid)
}
