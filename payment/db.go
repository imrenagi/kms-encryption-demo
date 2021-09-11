package payment

import (
  "context"

  "gorm.io/gorm"
)

type UserRepository struct {
  DB *gorm.DB
}

func (u *UserRepository) FindByID(ctx context.Context, ID string) (*User, error) {
  var user *User
  err := u.DB.WithContext(ctx).
    Where("users.id = ?", ID).
    Preload("CreditCards").
    First(&user).Error
  if err != nil {
    return nil, err
  }

  return user, nil
}

func (u *UserRepository) Save(ctx context.Context, user *User) error {
  return u.DB.WithContext(ctx).Create(user).Error
}

func (u *UserRepository) FindAll(ctx context.Context) ([]User, error) {
  var users []User
  err := u.DB.WithContext(ctx).
    Preload("CreditCards").
    Find(&users).Error
  if err != nil {
    return nil, err
  }
  return users, nil
}

func (u *UserRepository) Rotate(ctx context.Context) error {
  users, err := u.FindAll(ctx)
  if err != nil {
    return err
  }

  err = u.DB.Transaction(func(tx *gorm.DB) error {
    for _, user := range users {
      tx1 := tx.Model(&user).
        Session(&gorm.Session{FullSaveAssociations: true}).
        Updates(&user)
      if tx1.Error != nil {
        return tx1.Error
      }
    }
    return nil
  })
  if err != nil {
    return err
  }

  return nil
}
