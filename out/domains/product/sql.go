package product

import (
    "context"
    "github.com/rakateja/repogen/out/page"
    "github.com/rakateja/repogen/out/database"
    
)

type sqlRepository struct {
    db *database.PostgreSQL
}

func NewSQLRepository(db *database.PostgreSQL) Repository {
    return &sqlRepository{db}
}

const (
    selectQuery = `
        SELECT
            id, 
title, 
created_by, 
updated_by, 
created_at, 
updated_at
        FROM product
    `
    selectCount = `
        SELECT COUNT(id) FROM product
    `
    insertQuery = `
        INSERT INTO product (
            id, 
title, 
created_by, 
updated_by, 
created_at, 
updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6)
    `
    updateQuery = `
        UPDATE product
        SET
    `
    
    selectVariantQuery = `
        SELECT
            id, 
product_id, 
title, 
image, 
created_at
        FROM variant
    `
    
    selectImageQuery = `
        SELECT
            id, 
product_id, 
is_main, 
src, 
alt, 
created_at
        FROM image
    `
    
    
    insertVariantQuery = `
        INSERT INTO variant (
            id, 
product_id, 
title, 
image, 
created_at
        ) VALUES ($1, $2, $3, $4, $5)
    `
    
    insertImageQuery = `
        INSERT INTO image (
            id, 
product_id, 
is_main, 
src, 
alt, 
created_at
        ) VALUES ($1, $2, $3, $4, $5, $6)
    `
    
)

func (repo *sqlRepository) Store(ctx context.Context, entity Product) error {
    exist, err := repo.existByID(ctx, entity.Id)
    if err != nil {
        return err
    }
    if exist {
        return repo.update(ctx, entity)
    }
    return repo.insert(ctx, entity)
}

func (repo *sqlRepository) Count(ctx context.Context) (total int, err error) {
    if err := repo.db.Get(ctx, &total, selectCount); err != nil {
        return 0, err
    }
    return total, nil
}

func (repo *sqlRepository) FindByID(ctx context.Context, id string) (res Product, err error) {
    err = repo.db.Get(&res, selectQuery + " WHERE entity_id = " + id)
    if err != nil {
        return
    }
    
    var Variant []Variant
    err = repo.db.Select(ctx, &Variant, selectVariantQuery + " WHERE entity_id = " + id)
    if err != nil {
        return
    }
    
    var Image []Image
    err = repo.db.Select(ctx, &Image, selectImageQuery + " WHERE entity_id = " + id)
    if err != nil {
        return
    }
    
    res.VariantList = Variant
    res.ImageList = Image
    
    return
}

func (repo *sqlRepository) FindByIDs(ctx context.Context, ids []string) (res []Product, err error) {
    if err = repo.db.Select(&res, selectQuery + " WHERE entity_id IN (:ids)", map[string]interface{}{
        "ids": ids, 
    }); err != nil {
        return
    }
    return
}

func (repo *sqlRepository) FindPage(ctx context.Context, pageNum int, pageSize int) (res page.Page[Product], err error) {
    total, err := repo.Count(ctx)
    if err != nil {
        return
    }
    var items []Product
    offset := page.GetOffset(pageNum, pageSize)
    if err := repo.db.Select(ctx, &items, " LIMIT $1 OFFSET $2", pageSize, offset); err != nil {
        return res, err
    }
    return page.New[Product](items, total, pageNum, pageSize), nil
}

func (repo *sqlRepository) existByID(ctx context.Context, id string) (bool, error) {
    var total int
    if err := repo.db.Get(ctx, &total, " WHERE id = " + id); err != nil {
        return false, err
    }
    return total > 0, nil
}

func (repo *sqlRepository) insert(ctx context.Context, tx *sqlx.Tx, entity Product) error {
    res, err := tx.Execute(
        insertQuery,
        entity.Id,
        entity.Title,
        entity.CreatedBy,
        entity.UpdatedBy,
        entity.CreatedAt,
        entity.UpdatedAt,
        
    )
    rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected <= 0 {
		return errors.New("no rows affected")
	}
    return nil
}

func (repo *sqlRepository) update(ctx context.Context, entity Product) error {
    res, err := tx.Execute(
        updateQuery,
        entity.Id,
        entity.Title,
        entity.CreatedBy,
        entity.UpdatedBy,
        entity.CreatedAt,
        entity.UpdatedAt,
        
    )
    _, err = res.RowsAffected()
	if err != nil {
		return err
	}
    
    if err = repo.deleteMultiVariant(entity.ID); err != nil {
        return err
    }
    
    if err = repo.deleteMultiImage(entity.ID); err != nil {
        return err
    }
    
    
    if err = repo.insertMultiVariant({Variant variant product [{Id string id id} {ProductId string productId product_id} {Title string title title} {Image string image image} {CreatedAt time.Time createdAt created_at}]}); err != nil {
        return err
    }
    
    if err = repo.insertMultiImage({Image image product [{Id string id id} {ProductId string productId product_id} {IsMain bool isMain is_main} {Src string src src} {Alt string alt alt} {CreatedAt time.Time createdAt created_at}]}); err != nil {
        return err
    }
    
    return nil
}


func (repo *sqlRepository) deleteMultiVariant(id string) error {
    return nil
} 

func (repo *sqlRepository) deleteMultiImage(id string) error {
    return nil
} 



func (repo *sqlRepository) insertMultiVariant(ls []Variant) error {
    return nil
}

func (repo *sqlRepository) insertMultiImage(ls []Image) error {
    return nil
}


func (repo *sqlRepository) selectVariantByIDs(ids []string) (res []Variant, err error) {
    if err := repo.db.Select(&res, selectVariantQuery + " WHERE product_id IN (:ids)",
        map[string]interface{}{
            "ids": ids,
        },
    ); err != nil {
        return res, err
    }
    return res, nil
}

func (repo *sqlRepository) selectImageByIDs(ids []string) (res []Image, err error) {
    if err := repo.db.Select(&res, selectImageQuery + " WHERE product_id IN (:ids)",
        map[string]interface{}{
            "ids": ids,
        },
    ); err != nil {
        return res, err
    }
    return res, nil
}
