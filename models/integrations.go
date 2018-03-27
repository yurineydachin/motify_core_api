package models

type Integration struct {
	ID        uint64 `db:"id_integration"`
	Hash      string `db:"i_hash"`
	UpdatedAt string `db:"i_updated_at"`
	CreatedAt string `db:"i_created_at"`
}

func (i *Integration) ToArgs() map[string]interface{} {
	return map[string]interface{}{
		"id_integration": i.ID,
		"i_hash":         i.Hash,
	}
}
