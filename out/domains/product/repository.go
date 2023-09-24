package product

import (
    "context"
    "github.com/rakateja/repogen/out/page"
    
)

type Repository interface {
    Store(ctx context.Context, entity Product) error
    Count(ctx context.Context) (int, error)
    FindByID(ctx context.Context, id string) (Product, error)
    FindByIDs(ctx context.Context, ids []string) ([]Product, error)
    FindPage(ctx context.Context, pageNum int, limit int) (page.Page[Product], error)
}


