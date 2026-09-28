package domain

type ProductPatch struct {
	Name        *string
	Description *string
	PriceCents  *int
	ImageURL    *string
	Barcode     *string
	IsActive    *bool
}

func (p *Product) Apply(patch ProductPatch) error {
	if patch.Name != nil {
		p.Name = *patch.Name
	}
	if patch.Description != nil {
		p.Description = *patch.Description
	}
	if patch.PriceCents != nil {
		p.PriceCents = *patch.PriceCents
	}
	if patch.ImageURL != nil {
		p.ImageURL = *patch.ImageURL
	}
	if patch.Barcode != nil {
		p.Barcode = *patch.Barcode
	}
	if patch.IsActive != nil {
		p.IsActive = *patch.IsActive
	}
	return p.Validate()
}

type MachinePatch struct {
	Location *string
	Model    *string
}

func (m *Machine) Apply(patch MachinePatch) error {
	if patch.Location != nil {
		m.Location = *patch.Location
	}
	if patch.Model != nil {
		m.Model = *patch.Model
	}
	return m.Validate()
}
