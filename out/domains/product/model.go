package product

import (
    "time"
    "time"
    "time"
    "time"
    
)

type Product struct {
Id string `json:"id" db:"id"`
Title string `json:"title" db:"title"`
CreatedBy string `json:"createdBy" db:"created_by"`
UpdatedBy string `json:"updatedBy" db:"updated_by"`
CreatedAt time.Time `json:"createdAt" db:"created_at"`
UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`

}


type Variant struct {
Id string `json:"id" db:"id"`
ProductId string `json:"productId" db:"product_id"`
Title string `json:"title" db:"title"`
Image string `json:"image" db:"image"`
CreatedAt time.Time `json:"createdAt" db:"created_at"`

}

type Image struct {
Id string `json:"id" db:"id"`
ProductId string `json:"productId" db:"product_id"`
IsMain bool `json:"isMain" db:"is_main"`
Src string `json:"src" db:"src"`
Alt string `json:"alt" db:"alt"`
CreatedAt time.Time `json:"createdAt" db:"created_at"`

}
