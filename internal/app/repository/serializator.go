package repository

// import(
// 	"Lab1/internal/app/domain"
// )

type CreatePackage struct{
	Title string 
	ShortDescription string
	Description string 
	IsDeleted bool   
	Status string
	Price float64 
	DeviceCounting int 
}

type UpdatePackage struct {
	Title string 
	ShortDescription string
	Description string 
	Status string 
	Price float64 
	DeviceCounting int 
}	