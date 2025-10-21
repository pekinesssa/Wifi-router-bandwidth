package domain
import(

)

type User struct{
	ID uint `gorm:"primaryKey"`
	Login string `gorm:"type:varchar(100);not null;unique"`
	HashedPassword string `gorm:"column:hashed_password;type:varchar(255);not null"`
	IsModerator bool`gorm:"column:is_moderator;not null;default:false"`
	IsDeleted bool `gorm:"column:is_deleted;not null;default:false"`
}
