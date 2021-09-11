package payment

import (
  "encoding/json"
  "fmt"
  "strings"
  "time"
  "unicode/utf8"

  "github.com/google/uuid"
  "gorm.io/gorm"
)

func NewUserBuilder(name string) *Builder {
  return &Builder{
    User: &User{
      ID:        uuid.New(),
      CreatedAt: time.Now(),
      UpdatedAt: time.Now(),
      Name:      name,
    },
  }
}

type Builder struct {
  User *User
}

func (b *Builder) AddCard(number, expireAt, cva string) *Builder {
  b.User.AddCard(number, expireAt, cva)
  return b
}

func (b *Builder) Build() *User {
  return b.User
}

type User struct {
  ID          uuid.UUID      `gorm:"type:uuid;not null;primarykey" json:"id"`
  CreatedAt   time.Time      `json:"-"`
  UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"-"`
  DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
  Name        string         `gorm:"type:text" faker:"name" json:"name"`
  CreditCards []CreditCard `json:"credit_cards"`
}

func (u *User) AddCard(number, expireAt, cva string) error {
  u.CreditCards = append(u.CreditCards, CreditCard{
    ID:        uuid.New(),
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
    Number:    number,
    ExpireAt:  expireAt,
    CVA:       cva,
  })
  return nil
}

func (u *User) BeforeSave(tx *gorm.DB) (err error) {
  u.Name = encrypt([]byte(u.Name), []byte(u.ID.String()))
  return nil
}

func (u *User) AfterFind(tx *gorm.DB) error {
  u.Name = decrypt(u.Name, []byte(u.ID.String()))
  fmt.Println(u.Name, "After Find")
  return nil
}

type CreditCard struct {
  ID        uuid.UUID `gorm:"type:uuid;not null;primarykey" json:"-"`
  CreatedAt time.Time      `json:"created_at"`
  UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"-"`
  DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
  UserID    uuid.UUID      `json:"-"`
  Number    string         `gorm:"type:text" json:"-" faker:"cc_number"`
  ExpireAt  string         `gorm:"type:text" faker:"oneof: 10/22, 03/22, 05/23" json:"expire_at"`
  CVA       string         `gorm:"type:text" json:"-" faker:"oneof: 123, 432, 512"`
}

func (c *CreditCard) MarshalJSON() ([]byte, error) {
  type Alias CreditCard
  return json.Marshal(&struct {
    Number string `json:"number"`
    *Alias
  }{
    Number: c.GetCensoredNumber(),
    Alias:  (*Alias)(c),
  })
}

func (c CreditCard) GetCensoredNumber() string {
  cclen := len(c.Number)
  return strings.Repeat("*", utf8.RuneCountInString(c.Number[:cclen-4])) + c.Number[cclen-4:cclen]
}

func (c *CreditCard) BeforeSave(tx *gorm.DB) (err error) {
  c.Number = encrypt([]byte(c.Number), []byte(c.ID.String()))
  c.ExpireAt = encrypt([]byte(c.ExpireAt), []byte(c.ID.String()))
  c.CVA = encrypt([]byte(c.CVA), []byte(c.ID.String()))
  return nil
}

func (c *CreditCard) AfterFind(tx *gorm.DB) error {
  c.Number = decrypt(c.Number, []byte(c.ID.String()))
  c.ExpireAt = decrypt(c.ExpireAt, []byte(c.ID.String()))
  c.CVA = decrypt(c.CVA, []byte(c.ID.String()))
  return nil
}

